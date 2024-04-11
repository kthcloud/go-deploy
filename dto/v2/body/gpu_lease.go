package body

import "time"

type GpuLeaseGpuGroup struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

type GpuLeaseRead struct {
	ID         string `json:"id"`
	GpuGroupID string `json:"gpuGroupId"`
	Active     bool   `json:"active"`
	UserID     string `json:"userId"`
	// VmID is set when the lease is attached to a VM.
	VmID *string `json:"vmId,omitempty"`

	QueuePosition int     `json:"queuePosition"`
	LeaseDuration float64 `bson:"leaseDuration"`

	// ActivatedAt specifies the time when the lease was activated. This is the time the user first attached the GPU
	// or 1 day after the lease was created if the user did not attach the GPU.
	ActivatedAt *time.Time `json:"activatedAt,omitempty"`
	// AssignedAt specifies the time when the lease was assigned to the user.
	AssignedAt *time.Time `json:"assignedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	// ExpiresAt specifies the time when the lease will expire.
	// This is only present if the lease is active.
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	ExpiredAt *time.Time `json:"expiredAt,omitempty"`
}

type GpuLeaseCreate struct {
	// GpuGroupID is used to specify the GPU to lease.
	// As such, the lease does not specify which specific GPU to lease, but rather the type of GPU to lease.
	GpuGroupID string `json:"gpuGroupId" bson:"gpuGroupId" binding:"required"`
	// LeaseForever is used to specify whether the lease should be created forever.
	LeaseForever bool `json:"leaseForever" bson:"leaseForever"`
}

type GpuLeaseUpdate struct {
	// VmID is used to specify the VM to attach the lease to.
	//
	// - If specified, the lease will be attached to the VM.
	//
	// - If the lease is already attached to a VM, it will be detached from the current VM and attached to the new VM.
	//
	// - If the lease is not active, specifying a VM will activate the lease.
	//
	// - If the lease is not assigned, an error will be returned.
	VmID *string `json:"vmId,omitempty" bson:"vmId,omitempty" binding:"omitempty,uuid4"`
}

type GpuLeaseCreated struct {
	ID    string `json:"id"`
	JobID string `json:"jobId"`
}

type GpuLeaseUpdated struct {
	ID    string `json:"id"`
	JobID string `json:"jobId"`
}

type GpuLeaseDeleted struct {
	ID    string `json:"id"`
	JobID string `json:"jobId"`
}
