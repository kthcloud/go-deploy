package gpu_lease_repo

import "fmt"

var (
	// ErrGpuLeaseAlreadyExists is returned when a GPU lease already exists for a user.
	ErrGpuLeaseAlreadyExists = fmt.Errorf("gpu lease already exists")

	// ErrVmAlreadyAttached is returned when a VM is already attached to a GPU lease.
	ErrVmAlreadyAttached = fmt.Errorf("vm already attached")
)
