package base

import (
	configModels "go-deploy/models/config"
	storage_manager2 "go-deploy/models/sys/storage_manager"
	"go-deploy/pkg/config"
	"go-deploy/service/resources"
)

type StorageManagerContext struct {
	StorageManager *storage_manager2.StorageManager
	Zone           *configModels.DeploymentZone
	Generator      *resources.PublicGeneratorType
}

func NewStorageManagerBaseContext(id string) (*StorageManagerContext, error) {
	storageManager, err := storage_manager2.New().GetByID(id)
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
