package v1

import "fmt"

var (
	AuthInfoNotAvailableErr = fmt.Errorf("auth info not available")
	InternalError           = fmt.Errorf("internal error")
	NoAvailableGpuErr       = fmt.Errorf("no available GPUs")
	GpuNotAvailableErr      = fmt.Errorf("GPU not available")
	HostNotAvailableErr     = fmt.Errorf("host not available")
)
