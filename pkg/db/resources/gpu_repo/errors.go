package gpu_repo

import (
	"fmt"
)

var (
	// NotFoundErr is returned when a GPU is not found
	NotFoundErr = fmt.Errorf("gpu_repo not found")
	// AlreadyAttachedErr is returned when a GPU is already attached to a VM
	AlreadyAttachedErr = fmt.Errorf("gpu_repo is already attached to a vm")
)
