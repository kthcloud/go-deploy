package body

import "time"

type VmSnapshotRead struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created"`
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
