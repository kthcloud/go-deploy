package client

import (
	"errors"
	"fmt"
	configModels "go-deploy/models/config"
	deploymentModel "go-deploy/models/sys/deployment"
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
			//deploymentStore:  make(map[string]*deploymentModel.Deployment),
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

func (c *BaseClient[parent]) Deployment() *deploymentModel.Deployment {
	if c.deployment == nil {
		err := c.Fetch()
		if err != nil {
			if errors.Is(err, sErrors.DeploymentNotFoundErr) {
				return nil
			}

			utils.PrettyPrintError(err)
			return nil
		}
	}

	return c.deployment
}

func (c *BaseClient[parent]) ID() string {
	if c.id != "" {
		return c.id
	}

	if c.Deployment() != nil {
		return c.Deployment().ID
	}

	return ""
}

func (c *BaseClient[parent]) HasID() bool {
	return c.ID() != ""
}

func (c *BaseClient[parent]) Name() string {
	if c.name != "" {
		return c.name
	}

	if c.Deployment() != nil {
		return c.Deployment().Name
	}

	return ""
}

func (c *BaseClient[parent]) Zone() *configModels.DeploymentZone {
	return c.zone
}

func (c *BaseClient[parent]) Fetch() error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to fetch deployment in service client: %w", err)
	}

	var deployment *deploymentModel.Deployment
	if c.id != "" {
		var err error
		deployment, err = deploymentModel.New().GetByID(c.id)
		if err != nil {
			return makeError(err)
		}
	} else if c.name != "" {
		var err error
		deployment, err = deploymentModel.New().GetByName(c.name)
		if err != nil {
			return makeError(err)
		}
	}

	if deployment == nil {
		return makeError(sErrors.DeploymentNotFoundErr)
	}

	zone := config.Config.Deployment.GetZone(deployment.Zone)
	if zone == nil {
		return makeError(sErrors.ZoneNotFoundErr)
	}

	mainApp := deployment.GetMainApp()
	if mainApp == nil {
		return makeError(sErrors.MainAppNotFoundErr)
	}

	c.zone = zone
	c.MainApp = mainApp
	c.deployment = deployment
	c.id = deployment.ID
	c.name = deployment.Name
	c.UserID = deployment.OwnerID

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
	c.UserID = userID
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
