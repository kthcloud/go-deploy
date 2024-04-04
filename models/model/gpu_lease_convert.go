package model

import (
	"go-deploy/dto/v2/body"
)

// ToDTO converts a GpuLease to a body.GpuLeaseRead DTO.
func (g *GpuLease) ToDTO(queuePosition int) body.GpuLeaseRead {
	return body.GpuLeaseRead{
		ID:            g.ID,
		VmID:          g.VmID,
		GpuGroupID:    g.GpuGroupID,
		Active:        g.IsActive(),
		QueuePosition: queuePosition,
		ActivatedAt:   g.ActivatedAt,
		AssignedAt:    g.AssignedAt,
		ExpiredAt:     g.ExpiredAt,
		CreatedAt:     g.CreatedAt,
	}
}
