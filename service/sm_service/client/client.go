package client

import (
	"fmt"
	smModels "go-deploy/models/sys/sm"
	"go-deploy/service"
	sErrors "go-deploy/service/errors"
)

type BaseClient[parent any] struct {
	p *parent

	*service.Cache
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

func (c *BaseClient[parent]) SM(id, userID string, smc *smModels.Client) (*smModels.SM, error) {
	sm := c.GetSM(id)
	if sm != nil {
		return sm, nil
	}

	sm = c.GetSM(userID)
	if sm != nil {
		return sm, nil
	}

	return c.fetchSM(id, smc)
}

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
		return nil, makeError(sErrors.SmNotFoundErr)
	}

	c.StoreSM(sm)
	return sm, nil
}

func (c *BaseClient[parent]) WithAuth(auth *service.AuthInfo) *parent {
	c.Auth = auth
	return c.p
}
