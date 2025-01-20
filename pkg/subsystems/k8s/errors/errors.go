package errors

import "fmt"

var (
	// ErrIngressHostInUse is returned when the ingress host is already in use.
	ErrIngressHostInUse = fmt.Errorf("ingress host is already in use")

	// ErrNotFound is returned when a resource is not found.
	ErrNotFound = fmt.Errorf("resource not found")
)
