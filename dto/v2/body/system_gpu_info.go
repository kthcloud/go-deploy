package body

import (
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
	HostBase `json:",inline" bson:",inline" tstype:",extends"`
	GPUs     []GpuInfo `bson:"gpus" json:"gpus"`
}

type GpuInfo struct {
	Name        string `bson:"name" json:"name"`
	Slot        string `bson:"slot" json:"slot"`
	Vendor      string `bson:"vendor" json:"vendor"`
	VendorID    string `bson:"vendorId" json:"vendorId"`
	Bus         string `bson:"bus" json:"bus"`
	DeviceID    string `bson:"deviceId" json:"deviceId"`
	Passthrough bool   `bson:"passthrough" json:"passthrough"`
}
