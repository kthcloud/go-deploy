package base

import (
	"go-deploy/models/sys/deployment/storage_manager"
	"go-deploy/models/sys/enviroment"
	"go-deploy/pkg/conf"
	"go-deploy/service/resources"
)

type StorageManagerContext struct {
	StorageManager *storage_manager.StorageManager
	Zone           *enviroment.DeploymentZone
	CreateParams   *storage_manager.CreateParams
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

	zone := conf.Env.Deployment.GetZone(storageManager.Zone)
	if zone == nil {
		return nil, ZoneNotFoundErr
	}

	return &StorageManagerContext{
		StorageManager: storageManager,
		Zone:           zone,
		Generator:      resources.PublicGenerator().WithStorageManager(storageManager).WithDeploymentZone(zone),
	}, nil
}

func (c *StorageManagerContext) WithCreateParams(params *storage_manager.CreateParams) *StorageManagerContext {
	c.CreateParams = params
	return c
}
