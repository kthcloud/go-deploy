package client

import (
	"fmt"
	smModels "go-deploy/models/sys/sm"
	"go-deploy/service"
)

// BaseClient is the base client for all the subsystems client for SMs.
type BaseClient[parent any] struct {
	p *parent

	// Cache is used to cache the resources fetched inside the service.
	Cache *service.Cache

	// Auth is the authentication information for the client.
	Auth *service.AuthInfo
}

// NewBaseClient creates a new base client for SMs.
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

// SM returns the SM with the given ID.
// After a successful fetch, the SM will be cached.
func (c *BaseClient[parent]) SM(id, userID string, smc *smModels.Client) (*smModels.SM, error) {
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

// SMs returns a list of SMs.
// After a successful fetch, the SMs will be cached.
func (c *BaseClient[parent]) fetchSM(id string, smc *smModels.Client) (*smModels.SM, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to fetch sm in service client: %w", err)
	}

	if smc == nil {
		smc = smModels.New()
	}

	var sm *smModels.SM
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

// WithAuth sets the auth on the context.
// This is used to perform authorization checks.
func (c *BaseClient[parent]) WithAuth(auth *service.AuthInfo) *parent {
	c.Auth = auth
	return c.p
}
