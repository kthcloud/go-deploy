package model

import (
	"go-deploy/dto/v2/body"
	"go-deploy/utils"
	"time"
)

// ToDTO converts a GpuLease to a body.GpuLeaseRead DTO.
func (g *GpuLease) ToDTO(queuePosition int) body.GpuLeaseRead {
	var expiresAt *time.Time
	if g.ActivatedAt != nil {
		expiresAt = utils.PtrOf(g.ActivatedAt.Add(time.Hour * time.Duration(g.LeaseDuration)))
	}

	return body.GpuLeaseRead{
		ID:            g.ID,
		VmID:          g.VmID,
		GpuGroupID:    g.GpuGroupID,
		Active:        g.IsActive(),
		QueuePosition: queuePosition,
		ActivatedAt:   g.ActivatedAt,
		AssignedAt:    g.AssignedAt,
		ExpiresAt:     expiresAt,
		ExpiredAt:     g.ExpiredAt,
		CreatedAt:     g.CreatedAt,
	}
}

// FromDTO converts body.GpuLeaseUpdate DTO to GpuLeaseUpdateParams.
func (p GpuLeaseUpdateParams) FromDTO(dto *body.GpuLeaseUpdate) *GpuLeaseUpdateParams {
	// Only allow activations and not deactivations.
	var active *bool
	if dto.Active != nil && *dto.Active {
		active = dto.Active
	}

	return &GpuLeaseUpdateParams{
		Active: active,
	}
}
