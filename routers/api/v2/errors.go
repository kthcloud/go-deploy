package v2

import "fmt"

var (
	// ErrAuthInfoNotAvailable is returned when the auth info is not available
	ErrAuthInfoNotAvailable = fmt.Errorf("auth info not available")
	// ErrAuthInfoSetupFailed is returned when the auth info setup failed
	ErrAuthInfoSetupFailed = fmt.Errorf("auth info setup failed")
	// ErrInternal is returned when an internal error occurs.
	// This is a generic error that should not be returned in the API, but rather logged
	ErrInternal = fmt.Errorf("internal error")
	// ErrNoAvailableGpu is returned when there are no available GPUs.
	// This is normally caused by all GPUs being leased
	ErrNoAvailableGpu = fmt.Errorf("no available GPUs")
	// ErrGpuNotAvailable is returned when the GPU is not available.
	// This is normally caused by the host being down
	ErrGpuNotAvailable = fmt.Errorf("GPU not available")
	// ErrHostNotAvailable is returned when the host is not available.
	// This is normally caused by the host being down or disabled/in maintenance
	ErrHostNotAvailable = fmt.Errorf("host not available")
	// ErrVmToLargeForHost is returned when the VM is too large for the host.
	// This is normally caused by the VM requesting more resources than the host has available
	ErrVmToLargeForHost = fmt.Errorf("VM too large for host")
	// ErrVmNotReady is returned when the VM is not ready.
	// This is normally caused by the internal CloudStack VM not being created yet
	ErrVmNotReady = fmt.Errorf("VM not ready")
)

// MakeVmToLargeForHostErr creates a ErrVmToLargeForHost with the available CPU and RAM
func MakeVmToLargeForHostErr(cpuAvailable, ramAvailable int) error {
	return fmt.Errorf("%w. There is only %d CPU cores and %d GB RAM available on the host", ErrVmToLargeForHost, cpuAvailable, ramAvailable)
}
