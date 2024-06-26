package gpu_group_repo

import "fmt"

var (
	// GpuLeaseAlreadyExistsErr is returned when a GPU lease already exists for a user.
	GpuLeaseAlreadyExistsErr = fmt.Errorf("gpu lease already exists")
)
