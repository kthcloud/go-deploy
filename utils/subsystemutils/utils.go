package subsystemutils

import (
	"fmt"
	"go-deploy/pkg/conf"
)

func GetPrefixedName(name string) string {
	return fmt.Sprintf("%s%s", conf.Env.Deployment.Prefix, name)
}
