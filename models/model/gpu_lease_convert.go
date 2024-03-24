package model

import (
	"go-deploy/dto/v2/body"
)

// ToDTO converts a GpuLease to a body.GpuLeaseRead DTO.
func (g *GpuLease) ToDTO() body.GpuLeaseRead {
	return body.GpuLeaseRead{
		ID:      g.ID,
		VmID:    g.VmID,
		GpuName: g.GroupName,
		Active:  g.IsActive(),
		// TODO
		EstimatedAvailableAt: nil,
		// TODO
		ActivatedAt: nil,
		// TODO
		AssignedAt: nil,
		CreatedAt:  g.CreatedAt,
	}
}
