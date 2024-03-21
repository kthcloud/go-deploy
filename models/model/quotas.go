package model

import "go-deploy/dto/v1/body"

type Quotas struct {
	Deployments      int     `yaml:"deployments" structs:"deployments"`
	CpuCores         int     `yaml:"cpuCores" structs:"cpuCores"`
	RAM              int     `yaml:"ram" structs:"ram"`
	DiskSize         int     `yaml:"diskSize" structs:"diskSize"`
	Snapshots        int     `yaml:"snapshots" structs:"snapshots"`
	GpuLeaseDuration float64 `yaml:"gpuLeaseDuration" structs:"gpuLeaseDuration"` // in hours
}

// ToDTO converts a Quotas to a body.Quota DTO.
func (q *Quotas) ToDTO() body.Quota {
	return body.Quota{
		Deployments:      q.Deployments,
		CpuCores:         q.CpuCores,
		RAM:              q.RAM,
		DiskSize:         q.DiskSize,
		Snapshots:        q.Snapshots,
		GpuLeaseDuration: q.GpuLeaseDuration,
	}
}
