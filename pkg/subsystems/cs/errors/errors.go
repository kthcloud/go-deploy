package errors

import "fmt"

// PortInUseError is returned when the port is already in use.
type PortInUseError struct {
	Port int
}

// Error returns the reason for the port in use error.
func (e PortInUseError) Error() string {
	return fmt.Sprintf("port %d is already in use", e.Port)
}

// NewPortInUseError creates a new PortInUseError.
func NewPortInUseError(port int) PortInUseError {
	return PortInUseError{Port: port}
}

var (
	// NotFoundErr is returned when the resource is not found.
	NotFoundErr = fmt.Errorf("not found")
)
