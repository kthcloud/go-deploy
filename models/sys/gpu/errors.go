package gpu

import (
	"fmt"
)

var (
	// NotFoundErr is returned when a gpu is not found
	NotFoundErr = fmt.Errorf("gpu not found")
	// AlreadyAttachedErr is returned when a gpu is already attached to a vm
	AlreadyAttachedErr = fmt.Errorf("gpu is already attached to a vm")
)
