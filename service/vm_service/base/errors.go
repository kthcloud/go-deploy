package base

import "fmt"

var (
	VmDeletedErr    = fmt.Errorf("vm deleted")
	ZoneNotFoundErr = fmt.Errorf("zone not found")
)
