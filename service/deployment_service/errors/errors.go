package errors

import (
	"fmt"
)

var (
	DeploymentNotFoundErr = fmt.Errorf("deployment not found")
	DeploymentDeletedErr  = fmt.Errorf("deployment deleted")
	MainAppNotFoundErr    = fmt.Errorf("main app not found")

	StorageManagerDeletedErr = fmt.Errorf("storage manager deleted")

	ZoneNotFoundErr = fmt.Errorf("zone not found")

	CustomDomainInUseErr   = fmt.Errorf("custom domain is already in use")
	NonUniqueFieldErr      = fmt.Errorf("non unique field")
	InvalidTransferCodeErr = fmt.Errorf("invalid transfer code")
)
