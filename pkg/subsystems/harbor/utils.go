package harbor

import (
	"deploy-api-go/utils/subsystemutils"
	"fmt"
)

func getRobotFullName(name string) string {
	return fmt.Sprintf("robot$%s", getRobotName(name))
}

func getRobotName(name string) string {
	return fmt.Sprintf("%s+%s", subsystemutils.GetPrefixedName(name), name)
}

func int64Ptr(i int64) *int64 { return &i }