package model

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/kthcloud/go-deploy/dto/v2/body"
	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s/parsers/dra"
	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s/parsers/dra/nvidia"
	"github.com/kthcloud/go-deploy/utils"
	"go.mongodb.org/mongo-driver/bson"
)

var (
	ErrUnknownParameterImplType = errors.New("unknown underlying opaque GPU config type")
	ErrCouldNotInferDriver      = errors.New("could not infer driver")
)

// GpuClaim represents a DRA-style claim for one or more GPUs
// that can be requested and consumed by deployments or workloads.
type GpuClaim struct {
	ID   string `bson:"id"`
	Name string `bson:"name"`
	Zone string `bson:"zone"`

	// Requested contains all requested GPU configurations by key request.Name
	// this key is the key that users will request for
	Requested map[string]RequestedGpu `bson:"requested"`

	// Allocated contains the GPUs that have been successfully bound/allocated
	Allocated map[string][]AllocatedGpu `bson:"allocated,omitempty"`

	// Consumers are the workloads currently using this claim
	Consumers []GpuClaimConsumer `bson:"consumers,omitempty"`

	// Status reflects the reconciliation and/or lifecycle state
	Status *GpuClaimStatus `bson:"status,omitempty"`

	// LastError holds the last reconciliation or provisioning error
	LastError error `bson:"lastError,omitempty"`

	// The roles that are allowed to use this GpuClaim
	// Empty means all roles
	AllowedRoles []string `bson:"allowedRoles,omitempty"`

	Activities map[string]Activity `bson:"activities"`

	Subsystems GpuClaimSubsystems `bson:"subsystems"`

	CreatedAt time.Time  `bson:"createdAt"`
	UpdatedAt *time.Time `bson:"updatedAt,omitempty"`
}

// RequestAllocationMode defines how GPUs should be allocated.
type RequestAllocationMode string

const (
	RequestAllocationMode_None       RequestAllocationMode = ""
	RequestAllocationMode_All        RequestAllocationMode = "All"
	RequestAllocationMode_ExactCount RequestAllocationMode = "ExactCount"
)

type RequestedGpuCreate struct {
	RequestedGpu `mapstructure:",squash" json:",inline" bson:",inline"`
	Name         string `json:"name" bson:"name"`
}

// RequestedGpu describes the desired GPU configuration a workload requests.
type RequestedGpu struct {
	AllocationMode  RequestAllocationMode   `bson:"allocationMode"`
	Capacity        map[string]string       `bson:"capacity,omitempty"` // e.g. memory: "16Gi"
	Count           *int64                  `bson:"count,omitempty"`
	DeviceClassName string                  `bson:"deviceClassName"`
	Selectors       []string                `bson:"selectors,omitempty"`
	Config          *GpuDeviceConfiguration `bson:"config,omitempty"`
}

// GpuDeviceConfiguration holds the DRA opaque driver parameters and metadata.
type GpuDeviceConfiguration struct {
	Driver     string           `bson:"driver"`
	Parameters dra.OpaqueParams `bson:"parameters,omitempty"`
}

func (cfg GpuDeviceConfiguration) MarshalBSON() ([]byte, error) {
	doc := bson.M{
		"driver": cfg.Driver,
	}

	if cfg.Parameters != nil {
		// Marshal interface to BSON using its JSON form (to preserve structure)
		paramBytes, err := bson.Marshal(cfg.Parameters)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal parameters: %w", err)
		}

		var paramDoc bson.M
		if err := bson.Unmarshal(paramBytes, &paramDoc); err != nil {
			return nil, err
		}
		doc["parameters"] = paramDoc
	}

	return bson.Marshal(doc)
}

func (cfg *GpuDeviceConfiguration) UnmarshalBSON(data []byte) error {
	var raw bson.M
	if err := bson.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("failed to unmarshal raw data: %w", err)
	}

	// Extract driver first
	if d, ok := raw["driver"].(string); ok {
		cfg.Driver = d
	}

	// Nothing more to do if no parameters
	paramsRaw, ok := raw["parameters"]
	if !ok || paramsRaw == nil {
		return nil
	}

	// Marshal the inner parameters object for re-decoding
	paramBytes, err := bson.Marshal(paramsRaw)
	if err != nil {
		return fmt.Errorf("failed to re-marshal parameters: %w", err)
	}

	switch cfg.Driver {
	case "gpu.nvidia.com":
		var nvidia nvidia.GPUConfigParametersImpl
		if err := bson.Unmarshal(paramBytes, &nvidia); err != nil {
			return fmt.Errorf("failed to decode nvidia parameters: %w", err)
		}
		cfg.Parameters = nvidia

	default:
		var generic dra.OpaqueParams
		if err := bson.Unmarshal(paramBytes, &generic); err != nil {
			return fmt.Errorf("failed to decode generic parameters: %w", err)
		}
		cfg.Parameters = generic
	}

	return nil
}

// InferDriver attempts to infer the GPU driver based on the OpaqueParams implementation.
func (g GpuDeviceConfiguration) InferDriver() (string, error) {
	if g.Parameters != nil {
		switch g.Parameters.(type) {
		case nvidia.GPUConfigParametersImpl:
			return "gpu.nvidia.com", nil
		// If some time in the future we have
		// more impls we can add support for
		// other vendors here:
		// case amd.GPUConfigParametersImpl:
		//     return "gpu.amd.com", nil
		default:
			return "", ErrUnknownParameterImplType
		}
	}
	return "", ErrCouldNotInferDriver
}

// AllocatedGpu represents a concrete allocated GPU or GPU share.
type AllocatedGpu struct {
	Pool        string `bson:"pool,omitempty"`
	Device      string `bson:"device,omitempty"`
	ShareID     string `bson:"shareID,omitempty"`
	AdminAccess bool   `bson:"adminAccess,omitempty"`
}

// GpuClaimConsumer describes a workload (Pod/Deployment/etc.) consuming this GPU claim.
type GpuClaimConsumer struct {
	APIGroup string `bson:"apiGroup,omitempty"`
	Resource string `bson:"resource,omitempty"`
	Name     string `bson:"name,omitempty"`
	UID      string `bson:"uid,omitempty"`
}

type GpuClaimStatusPhase string

const (
	GpuClaimStatusPhase_Unknown GpuClaimStatusPhase = ""
	GpuClaimStatusPhase_Pending GpuClaimStatusPhase = "pending"
	GpuClaimStatusPhase_Bound   GpuClaimStatusPhase = "bound"
	GpuClaimStatusPhase_Failed  GpuClaimStatusPhase = "failed"
)

// GpuClaimStatus represents runtime state and metadata about allocation progress.
type GpuClaimStatus struct {
	Phase      GpuClaimStatusPhase `bson:"phase,omitempty"` // e.g. pending, bound, released, failed
	Message    string              `bson:"message,omitempty"`
	UpdatedAt  *time.Time          `bson:"updatedAt,omitempty"`
	LastSynced *time.Time          `bson:"lastSynced,omitempty"`
}

func (g GpuClaim) ToDTO() body.GpuClaimRead {
	dto := body.GpuClaimRead{
		ID:           g.ID,
		Name:         g.Name,
		Zone:         g.Zone,
		AllowedRoles: slices.Clone(g.AllowedRoles),
		CreatedAt:    g.CreatedAt,
		UpdatedAt:    g.UpdatedAt,
		LastError:    utils.ErrorStr(g.LastError),
	}

	// Convert Requested
	dto.Requested = make(map[string]body.RequestedGpu)
	for key, req := range g.Requested {
		dto.Requested[key] = body.RequestedGpu{
			AllocationMode:  string(req.AllocationMode),
			Capacity:        req.Capacity,
			Count:           req.Count,
			DeviceClassName: req.DeviceClassName,
			Selectors:       req.Selectors,
			Config: func() *body.GpuDeviceConfigurationWrapper {
				if req.Config == nil {
					return nil
				}
				if req.Config.Parameters == nil {
					driver := strings.TrimSpace(req.Config.Driver)
					if driver == "" {
						return nil
					}
					switch driver {
					case "gpu.nvidia.com":
						return &body.GpuDeviceConfigurationWrapper{
							GpuDeviceConfiguration: body.NvidiaDeviceConfiguration{
								Driver: driver,
							},
						}
					default:
						return &body.GpuDeviceConfigurationWrapper{
							GpuDeviceConfiguration: body.GenericDeviceConfiguration{
								Driver: driver,
							},
						}
					}
				}
				switch t := req.Config.Parameters.(type) {
				case nvidia.GPUConfigParametersImpl:
					return &body.GpuDeviceConfigurationWrapper{
						GpuDeviceConfiguration: body.NvidiaDeviceConfiguration{
							Driver:     req.Config.Driver,
							Parameters: &t.GpuConfig,
						},
					}
				default:
					return &body.GpuDeviceConfigurationWrapper{
						GpuDeviceConfiguration: body.GenericDeviceConfiguration{
							Driver: req.Config.Driver,
						},
					}
				}
			}(),
		}
	}

	// Convert Allocated
	dto.Allocated = make(map[string][]body.AllocatedGpu)
	for key, allocs := range g.Allocated {
		dto.Allocated[key] = make([]body.AllocatedGpu, len(allocs))
		for i, alloc := range allocs {
			dto.Allocated[key][i] = body.AllocatedGpu{
				Pool:        alloc.Pool,
				Device:      alloc.Device,
				ShareID:     alloc.ShareID,
				AdminAccess: alloc.AdminAccess,
			}
		}

	}

	// Convert Consumers
	for _, c := range g.Consumers {
		dto.Consumers = append(dto.Consumers, body.GpuClaimConsumer{
			APIGroup: c.APIGroup,
			Resource: c.Resource,
			Name:     c.Name,
			UID:      c.UID,
		})
	}

	// Convert Status
	if g.Status != nil {
		dto.Status = &body.GpuClaimStatus{
			Phase:      string(g.Status.Phase),
			Message:    g.Status.Message,
			UpdatedAt:  g.Status.UpdatedAt,
			LastSynced: g.Status.LastSynced,
		}
	}

	return dto
}

func (g GpuClaim) ToBriefDTO() body.GpuClaimRead {
	dto := body.GpuClaimRead{
		ID:        g.ID,
		Name:      g.Name,
		Zone:      g.Zone,
		CreatedAt: g.CreatedAt,
		UpdatedAt: g.UpdatedAt,
	}

	return dto
}

// DoingActivity returns true if the gpuClaim is doing the given activity.
func (gc *GpuClaim) DoingActivity(activity string) bool {
	for _, a := range gc.Activities {
		if a.Name == activity {
			return true
		}
	}
	return false
}

// Check if a user is allowed to use a gpuClaim
func (gc *GpuClaim) HasAccess(roles ...string) bool {
	if len(gc.AllowedRoles) == 0 {
		return true
	}
	for _, role := range roles {
		if slices.Contains(gc.AllowedRoles, role) {
			return true
		}
	}
	return false
}
