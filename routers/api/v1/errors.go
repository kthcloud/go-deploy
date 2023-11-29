package v1

import "fmt"

var (
	AuthInfoNotAvailableErr = fmt.Errorf("auth info not available")
	InternalError           = fmt.Errorf("internal error")
	NoAvailableGpuErr       = fmt.Errorf("no available GPUs")
	GpuNotAvailableErr      = fmt.Errorf("GPU not available")
	HostNotAvailableErr     = fmt.Errorf("host not available")
	VmTooLargeForHostErr    = fmt.Errorf("VM too large for host")
	VmNotReadyErr           = fmt.Errorf("VM not ready")
)

func MakeVmToLargeForHostErr(cpuAvailable, ramAvailable int) error {
	return fmt.Errorf("%w. CPU available: %d, RAM available: %d", VmTooLargeForHostErr, cpuAvailable, ramAvailable)
}
