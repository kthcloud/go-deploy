package body

import (
	"time"
)

type SystemStatus struct {
	Hosts []HostStatus `json:"hosts" bson:"hosts"`
}

type TimestampedSystemStatus struct {
	Status    SystemStatus `json:"status" bson:"status"`
	Timestamp time.Time    `json:"timestamp" bson:"timestamp"`
}

type HostStatus struct {
	HostBase `json:",inline" bson:",inline" tstype:",extends"`
	CPU      CpuStatus  `json:"cpu" bson:"cpu"`
	RAM      RamStatus  `json:"ram" bson:"ram"`
	GPU      *GpuStatus `json:"gpu,omitempty" bson:"gpu,omitempty"`
}

type CpuStatus struct {
	Temp CpuStatusTemp `json:"temp" bson:"temp"`
	Load CpuStatusLoad `json:"load" bson:"load"`
}

type CpuStatusTemp struct {
	Main  float64 `json:"main" bson:"main"`
	Cores []int   `json:"cores" bson:"cores"`
	Max   float64 `json:"max" bson:"max"`
}

type CpuStatusLoad struct {
	Main  float64 `json:"main" bson:"main"`
	Cores []int   `json:"cores" bson:"cores"`
	Max   float64 `json:"max" bson:"max"`
}

type RamStatus struct {
	Load RamStatusLoad `json:"load" bson:"load"`
}

type RamStatusLoad struct {
	Main float64 `json:"main" bson:"main"`
}

type GpuStatus struct {
	Temp []GpuStatusTemp `json:"temp" bson:"temp"`
}

type GpuStatusTemp struct {
	Main float64 `json:"main" bson:"main"`
}
