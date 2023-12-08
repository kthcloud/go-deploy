package base

import "fmt"

var (
	StorageManagerDeletedErr = fmt.Errorf("storage manager deleted")
	ZoneNotFoundErr          = fmt.Errorf("zone not found")
)
