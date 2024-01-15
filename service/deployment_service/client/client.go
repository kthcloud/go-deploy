package client

import (
	"fmt"
	deploymentModels "go-deploy/models/sys/deployment"
	"go-deploy/service"
)

// BaseClient is the base client for all the subsystems client for deployments.
type BaseClient[parent any] struct {
	p *parent

	// Cache is used to cache the resources fetched inside the service.
	Cache *service.Cache

	// Auth is the authentication information for the client.
	Auth *service.AuthInfo
}

// NewBaseClient creates a new BaseClient.
func NewBaseClient[parent any](cache *service.Cache) BaseClient[parent] {
	if cache == nil {
		cache = service.NewCache()
	}

	return BaseClient[parent]{Cache: cache}
}

// SetParent sets the parent of the client.
// This ensures the correct parent client is returned when calling builder methods.
func (c *BaseClient[parent]) SetParent(p *parent) {
	c.p = p
}

// Deployment returns the deployment with the given ID.
// After a successful fetch, the deployment is cached.
func (c *BaseClient[parent]) Deployment(id string, dmc *deploymentModels.Client) (*deploymentModels.Deployment, error) {
	deployment := c.Cache.GetDeployment(id)
	if deployment == nil {
		return c.fetchDeployment(id, "", dmc)
	}

	return deployment, nil
}

// Deployments returns a list of all deployments.
// After a successful fetch, the deployments are cached.
func (c *BaseClient[parent]) Deployments(dmc *deploymentModels.Client) ([]deploymentModels.Deployment, error) {
	// Right now we don't have a way to skip fetching when requesting a list of resources
	return c.fetchDeployments(dmc)
}

// WithAuth sets the auth on the context.
// This is used to perform authorization checks.
func (c *BaseClient[parent]) WithAuth(auth *service.AuthInfo) *parent {
	c.Auth = auth
	return c.p
}

// Refresh clears the cache for the deployment with the given ID and fetches it again.
// After a successful fetch, the deployment is cached.
func (c *BaseClient[parent]) Refresh(id string) (*deploymentModels.Deployment, error) {
	return c.fetchDeployment(id, "", nil)
}

// fetchDeployment fetches the deployment with the given ID or name.
// After a successful fetch, the deployment is cached.
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

// fetchDeployments fetches all deployments.
// After a successful fetch, the deployments are cached.
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
