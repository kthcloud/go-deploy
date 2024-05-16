package v1

import (
	"go-deploy/service/clients"
	"go-deploy/service/core"
	"go-deploy/service/v1/api"
	"go-deploy/service/v1/deployments"
	"go-deploy/service/v1/discovery"
	"go-deploy/service/v1/events"
	"go-deploy/service/v1/jobs"
	"go-deploy/service/v1/notifications"
	"go-deploy/service/v1/resource_migrations"
	"go-deploy/service/v1/sms"
	"go-deploy/service/v1/status"
	"go-deploy/service/v1/teams"
	"go-deploy/service/v1/users"
	"go-deploy/service/v1/zones"
)

type Client struct {
	V2 clients.V2

	auth  *core.AuthInfo
	cache *core.Cache
}

func New(v2 clients.V2, authInfo ...*core.AuthInfo) *Client {
	var auth *core.AuthInfo
	if len(authInfo) > 0 {
		auth = authInfo[0]
	}

	return &Client{
		V2:    v2,
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

func (c *Client) Deployments() api.Deployments {
	return deployments.New(c, c.cache)
}

func (c *Client) Discovery() api.Discovery {
	return discovery.New(c, c.cache)
}

func (c *Client) Events() api.Events {
	return events.New(c, c.cache)
}

func (c *Client) Jobs() api.Jobs {
	return jobs.New(c, c.cache)
}

func (c *Client) Notifications() api.Notifications {
	return notifications.New(c, c.cache)
}

func (c *Client) SMs() api.SMs {
	return sms.New(c, c.cache)
}

func (c *Client) Status() api.Status {
	return status.New(c, c.cache)
}

func (c *Client) Teams() api.Teams {
	return teams.New(c, c.cache)
}

func (c *Client) Users() api.Users {
	return users.New(c, c.V2, c.cache)
}

func (c *Client) Zones() api.Zones {
	return zones.New(c, c.cache)
}

func (c *Client) ResourceMigrations() api.ResourceMigrations {
	return resource_migrations.New(c, c.V2, c.cache)
}
