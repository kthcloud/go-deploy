package body

import "time"

type GpuAttached struct {
	ID    string `json:"id"`
	JobID string `json:"jobId"`
}

type GpuDetached struct {
	ID    string `json:"id"`
	JobID string `json:"jobId"`
}

type Lease struct {
	VmID    *string   `json:"vmId,omitempty"`
	User    *string   `json:"user,omitempty"`
	End     time.Time `json:"end"`
	Expired bool      `json:"expired"`
}

type GpuRead struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Zone string `json:"zone"`

	Lease *Lease `json:"lease,omitempty"`
}
