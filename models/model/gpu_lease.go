package model

import (
	"go-deploy/dto/v2/body"
	"time"
)

type GpuLease struct {
	ID         string `bson:"id"`
	GpuGroupID string `bson:"gpuGroupId"`
	UserID     string `bson:"userId"`

	// VmID is set to attach the lease to a VM.
	// If the lease is not attached to a VM, this field is nil.
	VmID *string `bson:"vmId"`

	LeaseDuration float64    `bson:"leaseDuration"`
	ActivatedAt   *time.Time `bson:"activatedAt,omitempty"`
	AssignedAt    *time.Time `bson:"assignedAt,omitempty"`
	ExpiredAt     *time.Time `bson:"expiredAt,omitempty"`
	CreatedAt     time.Time  `bson:"createdAt"`
}

// IsActive returns true if the lease is active.
// An active lease is subject to be expired.
func (g *GpuLease) IsActive() bool {
	return g.ActivatedAt != nil
}

// IsExpired returns true if the lease is expired.
func (g *GpuLease) IsExpired() bool {
	return g.ExpiredAt != nil && g.ExpiredAt.Before(time.Now())
}

type GpuLeaseCreateParams struct {
	GpuGroupName string
	LeaseForever bool
}

// FromDTO converts body.GpuLeaseCreate DTO to GpuLeaseCreateParams.
func (g GpuLeaseCreateParams) FromDTO(dto *body.GpuLeaseCreate) GpuLeaseCreateParams {
	return GpuLeaseCreateParams{
		GpuGroupName: dto.GpuGroupID,
		LeaseForever: dto.LeaseForever,
	}
}
