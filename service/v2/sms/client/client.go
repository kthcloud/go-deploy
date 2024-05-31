package client

import (
	"fmt"
	"go-deploy/models/model"
	"go-deploy/pkg/db/resources/sm_repo"
	"go-deploy/service/core"
)

// BaseClient is the base client for all the subsystems client for SMs.
type BaseClient[parent any] struct {
	p *parent

	// Cache is used to cache the resources fetched inside the service.
	Cache *core.Cache
}

// NewBaseClient creates a new base client for SMs.
func NewBaseClient[parent any](cache *core.Cache) BaseClient[parent] {
	if cache == nil {
		cache = core.NewCache()
	}

	return BaseClient[parent]{Cache: cache}
}

// SetParent sets the parent of the client.
// This ensures the correct parent client is returned when calling builder methods.
func (c *BaseClient[parent]) SetParent(p *parent) {
	c.p = p
}

// SM returns the SM with the given ID.
// After a successful fetch, the SM will be cached.
func (c *BaseClient[parent]) SM(id, userID string, smc *sm_repo.Client) (*model.SM, error) {
	sm := c.Cache.GetSM(id)
	if sm != nil {
		return sm, nil
	}

	sm = c.Cache.GetSM(userID)
	if sm != nil {
		return sm, nil
	}

	return c.fetchSM(id, smc)
}

// Refresh clears the cache for the SM with the given ID and fetches it again.
// After a successful fetch, the SM is cached.
func (c *BaseClient[parent]) Refresh(id string) (*model.SM, error) {
	return c.fetchSM(id, nil)
}

// SMs returns a list of SMs.
// After a successful fetch, the SMs will be cached.
func (c *BaseClient[parent]) fetchSM(id string, smc *sm_repo.Client) (*model.SM, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to fetch sm in service client: %w", err)
	}

	if smc == nil {
		smc = sm_repo.New()
	}

	var sm *model.SM
	var err error

	if id == "" {
		sm, err = smc.Get()
	} else {
		sm, err = smc.GetByID(id)
	}

	if err != nil {
		return nil, makeError(err)
	}

	if sm == nil {
		return nil, nil
	}

	c.Cache.StoreSM(sm)
	return sm, nil
}
