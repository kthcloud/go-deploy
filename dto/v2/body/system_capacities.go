package body

import (
	"time"
)

type TimestampedSystemCapacities struct {
	Capacities SystemCapacities `json:"capacities" bson:"capacities"`
	Timestamp  time.Time        `json:"timestamp" bson:"timestamp"`
}

type SystemCapacities struct {
	CpuCore CpuCoreCapacities `json:"cpuCore" bson:"cpuCore"`
	RAM     RamCapacities     `json:"ram" bson:"ram"`
	GPU     GpuCapacities     `json:"gpu" bson:"gpu"`
	Hosts   []HostCapacities  `json:"hosts" bson:"hosts"`
}

type ClusterCapacities struct {
	Name    string            `json:"cluster" bson:"cluster"`
	RAM     RamCapacities     `json:"ram" bson:"ram"`
	CpuCore CpuCoreCapacities `json:"cpuCore" bson:"cpuCore"`
}

type HostGpuCapacities struct {
	Count int `json:"count" bson:"count"`
}

type HostRamCapacities struct {
	Total int `json:"total" bson:"total"`
}

type HostCapacities struct {
	HostBase `json:",inline" bson:",inline" tstype:",extends"`
	CpuCore  CpuCoreCapacities `json:"cpuCore" bson:"cpuCore"`
	RAM      HostRamCapacities `json:"ram" bson:"ram"`
	GPU      HostGpuCapacities `json:"gpu" bson:"gpu"`
}

type RamCapacities struct {
	Total int `json:"total" bson:"total"`
}

type CpuCoreCapacities struct {
	Total int `json:"total" bson:"total"`
}

type GpuCapacities struct {
	Total int `json:"total" bson:"total"`
}
