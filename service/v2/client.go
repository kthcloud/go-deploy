package v2

import (
	"go-deploy/service/core"
)

type Client struct {
	auth  *core.AuthInfo
	cache *core.Cache
}

func New(authInfo ...*core.AuthInfo) *Client {
	var a *core.AuthInfo
	if len(authInfo) > 0 {
		a = authInfo[0]
	}

	return &Client{
		auth:  a,
		cache: core.NewCache(),
	}
}

func (c *Client) Auth() *core.AuthInfo {
	return c.auth
}

func (c *Client) HasAuth() bool {
	return c.auth != nil
}
