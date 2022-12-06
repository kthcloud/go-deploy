package k8s

import (
	"fmt"
	"go-deploy/pkg/conf"
	"go-deploy/utils/subsystemutils"
)

var manifestLabelName = "app.kubernetes.io/name"

func getDockerImageName(name string) string {
	projectName := subsystemutils.GetPrefixedName(name)
	return fmt.Sprintf("%s/%s/%s", conf.Env.DockerRegistry.Url, projectName, name)
}

func int32Ptr(i int32) *int32 { return &i }
