package gpu_lease_repo

import "fmt"

var (
	// GpuLeaseAlreadyExistsErr is returned when a GPU lease already exists for a user.
	GpuLeaseAlreadyExistsErr = fmt.Errorf("gpu lease already exists")

	// VmAlreadyAttachedErr is returned when a VM is already attached to a GPU lease.
	VmAlreadyAttachedErr = fmt.Errorf("vm already attached")
)
