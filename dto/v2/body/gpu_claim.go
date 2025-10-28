package body

import (
	"encoding/json"
	"time"

	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s/api/nvidia"
)

// GpuClaimAdminRead is a detailed DTO for administrators
// providing full visibility into requested, allocated,
// and consumed GPU resources.
type GpuClaimAdminRead struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Zone string `json:"zone"`

	// Requested contains all requested GPU configurations by key (request.Name).
	Requested map[string]RequestedGpu `json:"requested"`

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

// RequestedGpu describes the desired GPU configuration that was requested.
type RequestedGpu struct {
	AllocationMode  string                 `json:"allocationMode"`
	Capacity        map[string]string      `json:"capacity,omitempty"`
	Count           *int64                 `json:"count,omitempty"`
	DeviceClassName string                 `json:"deviceClassName"`
	Selectors       []string               `json:"selectors,omitempty"`
	Config          GpuDeviceConfiguration `json:"config,omitempty"`
}

// GpuDeviceConfiguration represents a vendor-specific GPU configuration.
type GpuDeviceConfiguration interface {
	DriverName() string
	json.Marshaler
}

// GenericDeviceConfiguration is a catch-all configuration when no vendor-specific struct is used.
type GenericDeviceConfiguration struct {
	Driver string `json:"driver,omitempty"`
}

func (g GenericDeviceConfiguration) DriverName() string {
	return g.Driver
}

func (g GenericDeviceConfiguration) MarshalJSON() ([]byte, error) {
	type Alias GenericDeviceConfiguration
	return json.Marshal(&struct {
		Type string `json:"type"`
		Alias
	}{
		Type:  "generic",
		Alias: (Alias)(g),
	})
}

// NvidiaDeviceConfiguration represents NVIDIA-specific configuration options.
type NvidiaDeviceConfiguration struct {
	Driver  string             `json:"driver,omitempty"`
	Sharing *nvidia.GpuSharing `json:"sharing,omitempty"`
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
	APIGroup  string `json:"apiGroup,omitempty"`
	Resource  string `json:"resource,omitempty"`
	Name      string `json:"name,omitempty"`
	UID       string `json:"uid,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

// GpuClaimStatus represents runtime state and metadata about allocation progress.
type GpuClaimStatus struct {
	Phase      string     `json:"phase,omitempty"`
	Message    string     `json:"message,omitempty"`
	UpdatedAt  *time.Time `json:"updatedAt,omitempty"`
	LastSynced *time.Time `json:"lastSynced,omitempty"`
}
