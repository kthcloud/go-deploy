package client

import (
	configModels "go-deploy/models/config"
	storageManagerModel "go-deploy/models/sys/storage_manager"
	"go-deploy/service"
	"go-deploy/service/resources"
)

type Context struct {
	id     string
	IDs    []string
	name   string
	userID string

	storageManager *storageManagerModel.StorageManager
	zone           *configModels.DeploymentZone
	Generator      *resources.PublicGeneratorType

	Auth *service.AuthInfo
}
