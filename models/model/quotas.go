package model

import (
	body "github.com/kthcloud/go-deploy/api/openapi/v2"
)

type Quotas struct {
	CpuCores         float64 `yaml:"cpuCores" structs:"cpuCores"`
	RAM              float64 `yaml:"ram" structs:"ram"`
	DiskSize         float64 `yaml:"diskSize" structs:"diskSize"`
	Snapshots        int     `yaml:"snapshots" structs:"snapshots"`
	GpuLeaseDuration float64 `yaml:"gpuLeaseDuration" structs:"gpuLeaseDuration"` // in hours
	Gpus             int     `yaml:"gpus" structs:"gpus"`
}

// ToDTO converts a Quotas to a body.Quota DTO.
func (q *Quotas) ToDTO() body.Quota {
	return dtoConverter.QuotasToDTO(q)
}
