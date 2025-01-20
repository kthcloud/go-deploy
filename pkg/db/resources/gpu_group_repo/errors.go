package gpu_group_repo

import "fmt"

var (
	// ErrGpuLeaseAlreadyExists is returned when a GPU lease already exists for a user.
	ErrGpuLeaseAlreadyExists = fmt.Errorf("gpu lease already exists")
)
