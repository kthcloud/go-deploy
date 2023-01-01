package subsystemutils

import (
	"fmt"
	"go-deploy/pkg/conf"
	"strings"
)

func GetPrefixedName(name string) string {
	return fmt.Sprintf("%s%s", conf.Env.AppPrefix, name)
}

func GetUnprefixedName(prefixedName string) string {
	return strings.TrimPrefix(prefixedName, conf.Env.AppPrefix)
}

func GetPlaceholderImage() (string, string) {
	result := strings.Split(conf.Env.DockerRegistry.PlaceHolder, ":")
	return result[0], result[1]
}
