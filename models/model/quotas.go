package model

import (
	"github.com/kthcloud/go-deploy/dto/v2/body"
)

type Quotas struct {
	CpuCores         float64 `yaml:"cpuCores" structs:"cpuCores"`
	RAM              float64 `yaml:"ram" structs:"ram"`
	DiskSize         float64 `yaml:"diskSize" structs:"diskSize"`
	Snapshots        int     `yaml:"snapshots" structs:"snapshots"`
	GpuLeaseDuration float64 `yaml:"gpuLeaseDuration" structs:"gpuLeaseDuration"` // in hours
}

// ToDTO converts a Quotas to a body.Quota DTO.
func (q *Quotas) ToDTO() body.Quota {
	return body.Quota{
		CpuCores:         q.CpuCores,
		RAM:              q.RAM,
		DiskSize:         q.DiskSize,
		Snapshots:        q.Snapshots,
		GpuLeaseDuration: q.GpuLeaseDuration,
	}
}
