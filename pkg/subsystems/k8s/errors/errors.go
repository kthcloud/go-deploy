package errors

import "fmt"

var (
	// IngressHostInUseErr is returned when the ingress host is already in use.
	IngressHostInUseErr = fmt.Errorf("ingress host is already in use")
)
