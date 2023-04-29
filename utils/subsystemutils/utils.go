package subsystemutils

import (
	"fmt"
	"go-deploy/pkg/conf"
	"strings"
)

func GetPrefixedName(name string) string {
	return fmt.Sprintf("%s%s", conf.Env.App.Prefix, name)
}

func GetUnprefixedName(prefixedName string) string {
	return strings.TrimPrefix(prefixedName, conf.Env.App.Prefix)
}

func GetPlaceholderImage() (string, string) {
	return conf.Env.DockerRegistry.Placeholder.Project, conf.Env.DockerRegistry.Placeholder.Repository
}
