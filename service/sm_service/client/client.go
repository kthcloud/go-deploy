package client

import (
	"fmt"
	smModels "go-deploy/models/sys/sm"
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
			smStore: make(map[string]*smModels.SM),
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

func (c *BaseClient[parent]) SM(id, userID string, smc *smModels.Client) (*smModels.SM, error) {
	sm, ok := c.smStore[id]
	if ok {
		return sm, nil
	}

	sm, ok = c.smStore[userID]
	if ok {
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

	c.storeSM(sm)
	return sm, nil
}

func (c *BaseClient[parent]) WithAuth(auth *service.AuthInfo) *parent {
	c.Auth = auth
	return c.p
}

func (c *BaseClient[parent]) storeSM(sm *smModels.SM) {
	c.smStore[sm.ID] = sm
	c.smStore[sm.OwnerID] = sm
}
