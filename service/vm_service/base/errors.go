package base

import "fmt"

var (
	DeploymentDeletedErr      = fmt.Errorf("deployment deleted")
	VmDeletedErr              = fmt.Errorf("vm deleted")
	ZoneNotFoundErr           = fmt.Errorf("zone not found")
	DeploymentZoneNotFoundErr = fmt.Errorf("deployment zone not found")
)
