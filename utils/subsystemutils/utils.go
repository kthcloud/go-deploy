package subsystemutils

import (
	"deploy-api-go/pkg/conf"
	"fmt"
	"strings"
)

func GetPrefixedName(name string) string {
	return fmt.Sprintf("%s%s", conf.Env.AppPrefix, name)
}

func GetPlaceholderImage() (string, string) {
	result := strings.Split(conf.Env.DockerRegistry.PlaceHolderImage, ":")
	return result[0], result[1]
}
