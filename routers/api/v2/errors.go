package v2

import "fmt"

var (
	// AuthInfoNotAvailableErr is returned when the auth info is not available
	AuthInfoNotAvailableErr = fmt.Errorf("auth info not available")
	// AuthInfoSetupFailedErr is returned when the auth info setup failed
	AuthInfoSetupFailedErr = fmt.Errorf("auth info setup failed")
	// InternalError is returned when an internal error occurs.
	// This is a generic error that should not be returned in the API, but rather logged
	InternalError = fmt.Errorf("internal error")
	// NoAvailableGpuErr is returned when there are no available GPUs.
	// This is normally caused by all GPUs being leased
	NoAvailableGpuErr = fmt.Errorf("no available GPUs")
	// GpuNotAvailableErr is returned when the GPU is not available.
	// This is normally caused by the host being down
	GpuNotAvailableErr = fmt.Errorf("GPU not available")
	// HostNotAvailableErr is returned when the host is not available.
	// This is normally caused by the host being down or disabled/in maintenance
	HostNotAvailableErr = fmt.Errorf("host not available")
	// VmTooLargeForHostErr is returned when the VM is too large for the host.
	// This is normally caused by the VM requesting more resources than the host has available
	VmTooLargeForHostErr = fmt.Errorf("VM too large for host")
	// VmNotReadyErr is returned when the VM is not ready.
	// This is normally caused by the internal CloudStack VM not being created yet
	VmNotReadyErr = fmt.Errorf("VM not ready")
)

// MakeVmToLargeForHostErr creates a VmTooLargeForHostErr with the available CPU and RAM
func MakeVmToLargeForHostErr(cpuAvailable, ramAvailable int) error {
	return fmt.Errorf("%w. There is only %d CPU cores and %d GB RAM available on the host", VmTooLargeForHostErr, cpuAvailable, ramAvailable)
}
