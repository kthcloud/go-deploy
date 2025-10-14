package model

import (
	"github.com/kthcloud/go-deploy/dto/v2/body"
	"github.com/kthcloud/go-deploy/utils"
	"time"
)

// ToDTO converts a GpuLease to a body.GpuLeaseRead DTO.
func (g *GpuLease) ToDTO(queuePosition int) body.GpuLeaseRead {
	var expiresAt *time.Time
	if g.ActivatedAt != nil {
		expiresAt = utils.PtrOf(g.ActivatedAt.Add(time.Hour * time.Duration(g.LeaseDuration)))
	}

	return body.GpuLeaseRead{
		ID:         g.ID,
		GpuGroupID: g.GpuGroupID,
		Active:     g.IsActive(),
		UserID:     g.UserID,
		VmID:       g.VmID,

		QueuePosition: queuePosition,
		LeaseDuration: g.LeaseDuration,

		ActivatedAt: g.ActivatedAt,
		AssignedAt:  g.AssignedAt,
		CreatedAt:   g.CreatedAt,
		ExpiresAt:   expiresAt,
		ExpiredAt:   g.ExpiredAt,
	}
}

// FromDTO converts body.GpuLeaseUpdate DTO to GpuLeaseUpdateParams.
func (p GpuLeaseUpdateParams) FromDTO(dto *body.GpuLeaseUpdate) *GpuLeaseUpdateParams {
	return &GpuLeaseUpdateParams{
		VmID: dto.VmID,
	}
}
