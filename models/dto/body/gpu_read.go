package body

import (
	"time"
)

type GpuLease struct {
	VmID string    `bson:"vmId" json:"vmId"`
	User string    `bson:"user" json:"user"`
	End  time.Time `bson:"end" json:"end"`
}

type GpuRead struct {
	ID    string    `json:"id"`
	Name  string    `json:"name"`
	Lease *GpuLease `json:"lease,omitempty"`
}
