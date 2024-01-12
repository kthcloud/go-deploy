package vm

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
