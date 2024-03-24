package body

import "time"

type GpuLeaseRead struct {
	ID      string `json:"id"`
	VmID    string `json:"vmId"`
	GpuName string `json:"gpuName"`
	Active  bool   `json:"active"`
	// EstimatedAvailableAt specifies the time when the GPU will be available if the leases in front in the
	// queue are not extended or released.
	EstimatedAvailableAt *time.Time `json:"estimatedAvailableAt,omitempty"`
	// ActivatedAt specifies the time when the lease was activated. This is the time the user first attached the GPU
	// or 1 day after the lease was created if the user did not attach the GPU.
	ActivatedAt *time.Time `json:"activatedAt,omitempty"`
	// AssignedAt specifies the time when the lease was assigned to the user.
	AssignedAt *time.Time `json:"assignedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
}

type GpuLeaseCreate struct {
	// GpuName is used to specify the GPU to lease.
	// As such, the lease does not specify which specific GPU to lease, but rather the type of GPU to lease.
	GpuName string `json:"gpuName" binding:"required"`
	// LeaseForever is used to specify whether the lease should be created forever.
	LeaseForever bool `json:"leaseForever"`
}

type GpuLeaseCreated struct {
	ID    string `json:"id"`
	JobID string `json:"jobId"`
}

type GpuLeaseDeleted struct {
	ID    string `json:"id"`
	JobID string `json:"jobId"`
}
