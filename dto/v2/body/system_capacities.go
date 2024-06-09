package body

import (
	"time"
)

type TimestampedSystemCapacities struct {
	Capacities SystemCapacities `json:"capacities" bson:"capacities"`
	Timestamp  time.Time        `json:"timestamp" bson:"timestamp"`
}

type SystemCapacities struct {
	// Total
	CpuCore CpuCoreCapacities `json:"cpuCore" bson:"cpuCore"`
	RAM     RamCapacities     `json:"ram" bson:"ram"`
	GPU     GpuCapacities     `json:"gpu" bson:"gpu"`

	// Per Host
	Hosts []HostCapacities `json:"hosts" bson:"hosts"`

	// Per Cluster
	Clusters []ClusterCapacities `json:"clusters" bson:"clusters"`
}

type ClusterCapacities struct {
	Name    string            `json:"cluster" bson:"cluster"`
	CpuCore CpuCoreCapacities `json:"cpuCore" bson:"cpuCore"`
	RAM     RamCapacities     `json:"ram" bson:"ram"`
	GPU     GpuCapacities     `json:"gpu" bson:"gpu"`
}

type HostCapacities struct {
	HostBase `json:",inline" bson:",inline" tstype:",extends"`
	CpuCore  CpuCoreCapacities `json:"cpuCore" bson:"cpuCore"`
	RAM      RamCapacities     `json:"ram" bson:"ram"`
	GPU      GpuCapacities     `json:"gpu" bson:"gpu"`
}

type CpuCoreCapacities struct {
	Total int `json:"total" bson:"total"`
}

type RamCapacities struct {
	Total int `json:"total" bson:"total"`
}

type GpuCapacities struct {
	Total int `json:"total" bson:"total"`
}
