package service_errors

import "fmt"

var (
	VmNotFoundErr          = fmt.Errorf("vm not found")
	VmNotCreatedErr        = fmt.Errorf("vm not created")
	NonUniqueFieldErr      = fmt.Errorf("non unique field")
	InvalidTransferCodeErr = fmt.Errorf("invalid transfer code")
	ZoneNotFoundErr        = fmt.Errorf("zone not found")
	GpuNotFoundErr         = fmt.Errorf("gpu not found")
	GpuAlreadyAttachedErr  = fmt.Errorf("gpu already attached")
	VmTooLargeErr          = fmt.Errorf("vm too large")
	HostNotAvailableErr    = fmt.Errorf("host not available")
)
