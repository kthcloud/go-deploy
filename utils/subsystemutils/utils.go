package subsystemutils

import (
	"fmt"
	"go-deploy/pkg/config"
)

func GetPrefixedName(name string) string {
	return fmt.Sprintf("%s%s", config.Config.Deployment.Prefix, name)
}
