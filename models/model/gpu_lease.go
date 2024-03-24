package model

import (
	"go-deploy/dto/v2/body"
	"time"
)

type GpuLease struct {
	ID        string `bson:"id"`
	GroupName string `bson:"groupName"`

	VmID   string `bson:"vmId"`
	UserID string `bson:"userId"`

	LeaseDuration float64    `bson:"leaseDuration"`
	ActivatedAt   *time.Time `bson:"expiresAt,omitempty"`

	CreatedAt time.Time `bson:"createdAt"`
}

// IsActive returns true if the lease is active.
// An active lease is subject to be expired.
func (g *GpuLease) IsActive() bool {
	return g.ActivatedAt != nil
}

// IsExpired returns true if the lease is expired.
func (g *GpuLease) IsExpired() bool {
	return g.ActivatedAt != nil && g.ActivatedAt.After(time.Now().Add(time.Duration(-g.LeaseDuration)*time.Hour))
}

type GpuLeaseCreateParams struct {
	GpuGroupName string
	LeaseForever bool
}

// FromDTO converts body.GpuLeaseCreate DTO to GpuLeaseCreateParams.
func (g GpuLeaseCreateParams) FromDTO(dto *body.GpuLeaseCreate) GpuLeaseCreateParams {
	return GpuLeaseCreateParams{
		GpuGroupName: dto.GpuName,
		LeaseForever: dto.LeaseForever,
	}
}
