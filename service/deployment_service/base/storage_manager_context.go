package base

import (
	configModels "go-deploy/models/config"
	"go-deploy/models/sys/deployment/storage_manager"
	"go-deploy/pkg/config"
	"go-deploy/service/resources"
)

type StorageManagerContext struct {
	StorageManager *storage_manager.StorageManager
	Zone           *configModels.DeploymentZone
	Generator      *resources.PublicGeneratorType
}

func NewStorageManagerBaseContext(id string) (*StorageManagerContext, error) {
	storageManager, err := storage_manager.New().GetByID(id)
	if err != nil {
		return nil, err
	}

	if storageManager == nil {
		return nil, StorageManagerDeletedErr
	}

	zone := config.Config.Deployment.GetZone(storageManager.Zone)
	if zone == nil {
		return nil, ZoneNotFoundErr
	}

	return &StorageManagerContext{
		StorageManager: storageManager,
		Zone:           zone,
		Generator:      resources.PublicGenerator().WithStorageManager(storageManager).WithDeploymentZone(zone),
	}, nil
}
