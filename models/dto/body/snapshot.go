package body

import "time"

type VmSnapshotRead struct {
	ID         string    `json:"id"`
	VmID       string    `json:"vmId"`
	Name       string    `json:"displayName"`
	ParentName *string   `json:"parentName,omitempty"`
	CreatedAt  time.Time `json:"created"`
	State      string    `json:"state"`
	Current    bool      `json:"current"`
}

type VmSnapshotCreate struct {
	Name string `json:"name" binding:"required,rfc1035,min=3,max=30"`
}

type VmSnapshotCreated struct {
	ID    string `json:"id"`
	JobID string `json:"jobId"`
}

type VmSnapshotDeleted struct {
	ID    string `json:"id"`
	JobID string `json:"jobId"`
}
