package harbor

import (
	"fmt"
	"go-deploy/utils/subsystemutils"
)

func getRobotFullName(name string) string {
	return fmt.Sprintf("robot$%s", getRobotName(name))
}

func getRobotName(name string) string {
	return fmt.Sprintf("%s+%s", subsystemutils.GetPrefixedName(name), name)
}

func int64Ptr(i int64) *int64 { return &i }
