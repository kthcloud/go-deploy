package client

import (
	"errors"
	"fmt"
	configModels "go-deploy/models/config"
	storageManagerModel "go-deploy/models/sys/storage_manager"
	"go-deploy/pkg/config"
	"go-deploy/service"
	sErrors "go-deploy/service/errors"
	"go-deploy/utils"
)

type BaseClient[parent any] struct {
	p *parent

	*Context
}

func NewBaseClient[parent any](context *Context) BaseClient[parent] {
	if context == nil {
		context = &Context{
			//smStore:  make(map[string]*storageManagerModel.StorageManager),
		}
	}

	return BaseClient[parent]{Context: context}
}

func (c *BaseClient[parent]) SetParent(p *parent) {
	c.p = p
}

func (c *BaseClient[parent]) SetContext(context *Context) {
	if context == nil {
		context = &Context{}
	}

	c.Context = context
}

func (c *BaseClient[parent]) StorageManager() *storageManagerModel.StorageManager {
	if c.storageManager == nil {
		err := c.Fetch()
		if err != nil {
			if errors.Is(err, sErrors.StorageManagerNotFoundErr) {
				return nil
			}

			utils.PrettyPrintError(err)
			return nil
		}
	}

	return c.storageManager
}

func (c *BaseClient[parent]) ID() string {
	if c.id != "" {
		return c.id
	}

	if c.StorageManager() != nil {
		return c.StorageManager().ID
	}

	return ""
}

func (c *BaseClient[parent]) HasID() bool {
	return c.ID() != ""
}

func (c *BaseClient[parent]) UserID() string {
	if c.userID != "" {
		return c.userID
	}

	if sm := c.StorageManager(); sm != nil {
		return sm.OwnerID
	}

	return ""
}

func (c *BaseClient[parent]) HasUserID() bool {
	return c.UserID() != ""
}

func (c *BaseClient[parent]) Zone() *configModels.DeploymentZone {
	return c.zone
}

func (c *BaseClient[parent]) Fetch() error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to fetch sm in service client: %w", err)
	}

	var storageManager *storageManagerModel.StorageManager
	if c.id != "" {
		var err error
		storageManager, err = storageManagerModel.New().GetByID(c.id)
		if err != nil {
			return makeError(err)
		}
	} else if c.name != "" {
		var err error
		storageManager, err = storageManagerModel.New().GetByName(c.name)
		if err != nil {
			return makeError(err)
		}
	}

	if storageManager == nil {
		return makeError(sErrors.StorageManagerNotFoundErr)
	}

	zone := config.Config.Deployment.GetZone(storageManager.Zone)
	if zone == nil {
		return makeError(sErrors.ZoneNotFoundErr)
	}

	c.zone = zone
	c.storageManager = storageManager
	c.id = storageManager.ID
	c.userID = storageManager.OwnerID

	return nil
}

func (c *BaseClient[parent]) WithID(id string) *parent {
	c.id = id
	return c.p
}

func (c *BaseClient[parent]) WithIDs(ids []string) *parent {
	c.IDs = ids
	return c.p
}

func (c *BaseClient[parent]) WithName(name string) *parent {
	c.name = name
	return c.p
}

func (c *BaseClient[parent]) WithUserID(userID string) *parent {
	c.userID = userID
	return c.p
}

func (c *BaseClient[parent]) WithZone(zone string) *parent {
	c.zone = config.Config.Deployment.GetZone(zone)
	return c.p
}

func (c *BaseClient[parent]) WithAuth(auth *service.AuthInfo) *parent {
	c.Auth = auth
	return c.p
}
