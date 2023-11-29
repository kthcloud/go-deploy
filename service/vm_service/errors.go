package vm_service

import "fmt"

var (
	VmNotFoundErr          = fmt.Errorf("vm not found")
	NonUniqueFieldErr      = fmt.Errorf("non unique field")
	InvalidTransferCodeErr = fmt.Errorf("invalid transfer code")
	GpuNotFoundErr         = fmt.Errorf("gpu not found")
	GpuAlreadyAttachedErr  = fmt.Errorf("gpu already attached")
)
