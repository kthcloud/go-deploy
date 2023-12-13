package client

import (
	"fmt"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/pkg/config"
	"go-deploy/service"
	sErrors "go-deploy/service/errors"
)

type BaseClient[parent any] struct {
	p *parent

	*Context
}

func NewBaseClient[parent any](context *Context) BaseClient[parent] {
	if context == nil {
		context = &Context{
			deploymentStore: make(map[string]*deploymentModel.Deployment),
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

func (c *BaseClient[parent]) Deployment(id string, dmc *deploymentModel.Client) (*deploymentModel.Deployment, error) {
	deployment, ok := c.deploymentStore[id]
	if !ok || deployment == nil {
		return c.fetchDeployment(id, "", dmc)
	}

	return deployment, nil
}

func (c *BaseClient[parent]) Deployments(dmc *deploymentModel.Client) ([]deploymentModel.Deployment, error) {
	// Right now we don't have a way to skip fetching when requesting a list of resources
	return c.fetchDeployments(dmc)
}

func (c *BaseClient[parent]) WithAuth(auth *service.AuthInfo) *parent {
	c.Auth = auth
	return c.p
}

func (c *BaseClient[parent]) Refresh(id string) (*deploymentModel.Deployment, error) {
	return c.fetchDeployment(id, "", nil)
}

func (c *BaseClient[parent]) fetchDeployment(id, name string, dmc *deploymentModel.Client) (*deploymentModel.Deployment, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to fetch deployment in service client: %w", err)
	}

	if dmc == nil {
		dmc = deploymentModel.New()
	}

	var deployment *deploymentModel.Deployment
	if id != "" {
		var err error
		deployment, err = dmc.GetByID(id)
		if err != nil {
			return nil, makeError(err)
		}
	} else if name != "" {
		var err error
		deployment, err = dmc.GetByName(name)
		if err != nil {
			return nil, makeError(err)
		}
	}

	if deployment == nil {
		return nil, nil
	}

	zone := config.Config.Deployment.GetZone(deployment.Zone)
	if zone == nil {
		return nil, makeError(sErrors.ZoneNotFoundErr)
	}

	return deployment, nil
}

func (c *BaseClient[parent]) fetchDeployments(dmc *deploymentModel.Client) ([]deploymentModel.Deployment, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to fetch gpus in service client: %w", err)
	}

	if dmc == nil {
		dmc = deploymentModel.New()
	}

	deployments, err := dmc.List()
	if err != nil {
		return nil, makeError(err)
	}

	for _, deployment := range deployments {
		c.storeDeployment(&deployment)
	}

	return deployments, nil
}

func (c *BaseClient[parent]) storeDeployment(deployment *deploymentModel.Deployment) {
	c.deploymentStore[deployment.ID] = deployment
}
