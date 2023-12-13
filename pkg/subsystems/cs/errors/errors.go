package errors

import "fmt"

var (
	// PortInUseErr is returned when a port is in use.
	PortInUseErr = fmt.Errorf("port is in use")
)
