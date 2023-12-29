package client

import (
	"fmt"
	deploymentModels "go-deploy/models/sys/deployment"
	"go-deploy/service"
)

type BaseClient[parent any] struct {
	p *parent

	// Cache is used to cache the resources fetched inside the service.
	Cache *service.Cache

	// Auth is the authentication information for the client.
	Auth *service.AuthInfo
}

func NewBaseClient[parent any](cache *service.Cache) BaseClient[parent] {
	if cache == nil {
		cache = service.NewCache()
	}

	return BaseClient[parent]{Cache: cache}
}

func (c *BaseClient[parent]) SetParent(p *parent) {
	c.p = p
}

func (c *BaseClient[parent]) Deployment(id string, dmc *deploymentModels.Client) (*deploymentModels.Deployment, error) {
	deployment := c.Cache.GetDeployment(id)
	if deployment == nil {
		return c.fetchDeployment(id, "", dmc)
	}

	return deployment, nil
}

func (c *BaseClient[parent]) Deployments(dmc *deploymentModels.Client) ([]deploymentModels.Deployment, error) {
	// Right now we don't have a way to skip fetching when requesting a list of resources
	return c.fetchDeployments(dmc)
}

func (c *BaseClient[parent]) WithAuth(auth *service.AuthInfo) *parent {
	c.Auth = auth
	return c.p
}

func (c *BaseClient[parent]) Refresh(id string) (*deploymentModels.Deployment, error) {
	return c.fetchDeployment(id, "", nil)
}

func (c *BaseClient[parent]) fetchDeployment(id, name string, dmc *deploymentModels.Client) (*deploymentModels.Deployment, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to fetch deployment in service client: %w", err)
	}

	if dmc == nil {
		dmc = deploymentModels.New()
	}

	var deployment *deploymentModels.Deployment
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

	c.Cache.StoreDeployment(deployment)

	return deployment, nil
}

func (c *BaseClient[parent]) fetchDeployments(dmc *deploymentModels.Client) ([]deploymentModels.Deployment, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to fetch gpus in service client: %w", err)
	}

	if dmc == nil {
		dmc = deploymentModels.New()
	}

	deployments, err := dmc.List()
	if err != nil {
		return nil, makeError(err)
	}

	for _, deployment := range deployments {
		d := deployment
		c.Cache.StoreDeployment(&d)
	}

	return deployments, nil
}
