package body

import (
	"go-deploy/pkg/subsystems/host_api"
	"time"
)

type SystemGpuInfo struct {
	HostGpuInfo []HostGpuInfo `json:"hosts" bson:"hosts"`
}

type TimestampedSystemGpuInfo struct {
	GpuInfo   SystemGpuInfo `json:"gpuInfo" bson:"gpuInfo"`
	Timestamp time.Time     `json:"timestamp" bson:"timestamp"`
}

type HostGpuInfo struct {
	HostBase `json:",inline" bson:",inline"`
	GPUs     []host_api.GpuInfo `bson:"gpus" json:"gpus"`
}
