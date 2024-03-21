package model

import "time"

type Snapshot struct {
	ID         string
	VmID       string
	Name       string
	ParentName *string
	CreatedAt  time.Time
	State      string
	Current    bool
}

type SnapshotV2 struct {
	ID        string
	Name      string
	Status    string
	CreatedAt time.Time
}
