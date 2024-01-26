package v2

import (
	"go-deploy/service/clients"
	"go-deploy/service/core"
	"go-deploy/service/v2/api"
	"go-deploy/service/v2/vms"
)

type Client struct {
	V1 clients.V1

	auth  *core.AuthInfo
	cache *core.Cache
}

func New(v1 clients.V1, authInfo ...*core.AuthInfo) *Client {
	var auth *core.AuthInfo
	if len(authInfo) > 0 {
		auth = authInfo[0]
	}

	return &Client{
		V1:    v1,
		auth:  auth,
		cache: core.NewCache(),
	}
}

func (c *Client) Auth() *core.AuthInfo {
	return c.auth
}

func (c *Client) HasAuth() bool {
	return c.auth != nil
}

func (c *Client) VMs() api.VMs {
	return vms.New(c.V1, c, c.cache)
}
