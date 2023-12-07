package errors

import (
	"fmt"
)

var (
	DeploymentNotFoundErr  = fmt.Errorf("deployment not found")
	ZoneNotFoundErr        = fmt.Errorf("zone not found")
	NonUniqueFieldErr      = fmt.Errorf("non unique field")
	InvalidTransferCodeErr = fmt.Errorf("invalid transfer code")
)
