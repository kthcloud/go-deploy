package base

import "fmt"

var (
	DeploymentDeletedErr     = fmt.Errorf("deployment deleted")
	StorageManagerDeletedErr = fmt.Errorf("storage manager deleted")
	MainAppNotFoundErr       = fmt.Errorf("main app not found")
	ZoneNotFoundErr          = fmt.Errorf("zone not found")
	CustomDomainInUseErr     = fmt.Errorf("custom domain is already in use")
)
