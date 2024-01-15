package subsystemutils

import (
	"fmt"
	"go-deploy/pkg/config"
)

// GetPrefixedName returns the name with the prefix.
// This is used to create names for resources.
func GetPrefixedName(name string) string {
	return fmt.Sprintf("%s%s", config.Config.Deployment.Prefix, name)
}
