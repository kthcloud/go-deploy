package errors

import "fmt"

var (
	// IngressHostInUseErr is returned when the ingress host is already in use.
	IngressHostInUseErr = fmt.Errorf("ingress host is already in use")

	// NotFoundErr is returned when a resource is not found.
	NotFoundErr = fmt.Errorf("resource not found")
)
