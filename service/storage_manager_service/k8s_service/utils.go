package k8s_service

import (
	"fmt"
	"go-deploy/models/config"
	storageManagerModel "go-deploy/models/sys/storage_manager"
	"path"
)

func GetStorageManagerVolumes(ownerID, appName string) ([]storageManagerModel.Volume, []storageManagerModel.Volume) {
	initVolumes := []storageManagerModel.Volume{
		{
			Name:    fmt.Sprintf("%s-%s", appName, "init"),
			Init:    false,
			AppPath: "/exports",
		},
	}

	volumes := []storageManagerModel.Volume{
		{
			Name:       fmt.Sprintf("%s-%s", appName, "data"),
			Init:       false,
			AppPath:    "/data",
			ServerPath: path.Join(ownerID, "data"),
		},
		{
			Name:       fmt.Sprintf("%s-%s", appName, "user"),
			Init:       false,
			AppPath:    "/deploy",
			ServerPath: path.Join(ownerID, "user"),
		},
	}

	return initVolumes, volumes
}

func GetExternalFQDN(name string, zone *config.DeploymentZone) string {
	return fmt.Sprintf("%s.%s", name, zone.ParentDomain)
}

func GetStorageManagerExternalFQDN(name string, zone *config.DeploymentZone) string {
	return fmt.Sprintf("%s.%s", name, zone.Storage.ParentDomain)
}
