package body

import (
	"encoding/json"
	"time"

	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s/api/nvidia"
)

// GpuClaimRead is a detailed DTO for administrators
// providing full visibility into requested, allocated,
// and consumed GPU resources.
type GpuClaimRead struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Zone string `json:"zone"`

	// Requested contains all requested GPU configurations by key (request.Name).
	Requested map[string]RequestedGpu `json:"requested,omitempty"`

	// Allocated contains the GPUs that have been successfully bound/allocated.
	Allocated map[string]AllocatedGpu `json:"allocated,omitempty"`

	// Consumers are the workloads currently using this claim.
	Consumers []GpuClaimConsumer `json:"consumers,omitempty"`

	// Status reflects the reconciliation and/or lifecycle state.
	Status *GpuClaimStatus `json:"status,omitempty"`

	// LastError holds the last reconciliation or provisioning error message.
	LastError string `json:"lastError,omitempty"`

	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
}

type GpuClaimCreate struct {
	Name string  `json:"name" bson:"name" binding:"required,rfc1035,min=3,max=30"`
	Zone *string `json:"zone" bson:"zone"`

	// Requested contains all requested GPU configurations by key (request.Name).
	Requested []RequestedGpuCreate `json:"requested,omitempty" bson:"requested,omitempty" binding:"min=1,dive"`
}

type GpuClaimCreated struct {
	ID    string `json:"id"`
	JobID string `json:"jobId"`
}

type RequestedGpuCreate struct {
	RequestedGpu `mapstructure:",squash" json:",inline" bson:",inline" binding:"required"`
	Name         string `json:"name" bson:"name" binding:"required,rfc1035,min=3,max=30"`
}

// RequestedGpu describes the desired GPU configuration that was requested.
type RequestedGpu struct {
	AllocationMode  string                 `json:"allocationMode" bson:"allocationMode" binding:"required,oneof=All ExactCount"`
	Capacity        map[string]string      `json:"capacity,omitempty" bson:"capacity,omitempty"`
	Count           *int64                 `json:"count,omitempty" bson:"count,omitempty"`
	DeviceClassName string                 `json:"deviceClassName" bson:"deviceClassName" binding:"required,rfc1123"`
	Selectors       []string               `json:"selectors,omitempty" bson:"selectors,omitempty"`
	Config          GpuDeviceConfiguration `json:"config,omitempty" bson:"config,omitempty"`
}

// GpuDeviceConfiguration represents a vendor-specific GPU configuration.
type GpuDeviceConfiguration interface {
	DriverName() string
	json.Marshaler
}

// GenericDeviceConfiguration is a catch-all configuration when no vendor-specific struct is used.
type GenericDeviceConfiguration struct {
	Driver string `json:"driver" bson:"driver"`
}

func (g GenericDeviceConfiguration) DriverName() string {
	return g.Driver
}

func (g GenericDeviceConfiguration) MarshalJSON() ([]byte, error) {
	type Alias GenericDeviceConfiguration
	return json.Marshal(&struct {
		Type string `json:"type" bson:"type"`
		Alias
	}{
		Type:  "generic",
		Alias: (Alias)(g),
	})
}

// NvidiaDeviceConfiguration represents NVIDIA-specific configuration options.
type NvidiaDeviceConfiguration struct {
	Driver  string             `json:"driver" bson:"driver"`
	Sharing *nvidia.GpuSharing `json:"sharing,omitempty" bson:"sharing,omitempty"`
}

func (NvidiaDeviceConfiguration) DriverName() string {
	return "gpu.nvidia.com"
}

func (n NvidiaDeviceConfiguration) MarshalJSON() ([]byte, error) {
	type Alias NvidiaDeviceConfiguration
	return json.Marshal(&struct {
		Type string `json:"type"`
		Alias
	}{
		Type:  "nvidia",
		Alias: (Alias)(n),
	})
}

// AllocatedGpu represents a concrete allocated GPU or GPU share.
type AllocatedGpu struct {
	Pool        string `json:"pool,omitempty"`
	Device      string `json:"device,omitempty"`
	ShareID     string `json:"shareID,omitempty"`
	AdminAccess bool   `json:"adminAccess,omitempty"`
}

// GpuClaimConsumer describes a workload consuming this GPU claim.
type GpuClaimConsumer struct {
	APIGroup string `json:"apiGroup,omitempty"`
	Resource string `json:"resource,omitempty"`
	Name     string `json:"name,omitempty"`
	UID      string `json:"uid,omitempty"`
}

// GpuClaimStatus represents runtime state and metadata about allocation progress.
type GpuClaimStatus struct {
	Phase      string     `json:"phase,omitempty"`
	Message    string     `json:"message,omitempty"`
	UpdatedAt  *time.Time `json:"updatedAt,omitempty"`
	LastSynced *time.Time `json:"lastSynced,omitempty"`
}
