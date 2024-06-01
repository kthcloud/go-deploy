package v2

import (
	"go-deploy/service/core"
	"go-deploy/service/v2/api"
	"go-deploy/service/v2/deployments"
	"go-deploy/service/v2/discovery"
	"go-deploy/service/v2/events"
	"go-deploy/service/v2/jobs"
	"go-deploy/service/v2/notifications"
	"go-deploy/service/v2/resource_migrations"
	"go-deploy/service/v2/sms"
	"go-deploy/service/v2/system"
	"go-deploy/service/v2/teams"
	"go-deploy/service/v2/users"
	"go-deploy/service/v2/vms"
)

type Client struct {
	auth  *core.AuthInfo
	cache *core.Cache
}

func New(authInfo ...*core.AuthInfo) *Client {
	var auth *core.AuthInfo
	if len(authInfo) > 0 {
		auth = authInfo[0]
	}

	return &Client{
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

func (c *Client) ResourceMigrations() api.ResourceMigrations {
	return resource_migrations.New(c, c.cache)
}

func (c *Client) SMs() api.SMs {
	return sms.New(c, c.cache)
}

func (c *Client) System() api.System {
	return system.New(c, c.cache)
}

func (c *Client) Teams() api.Teams {
	return teams.New(c, c.cache)
}

func (c *Client) Users() api.Users {
	return users.New(c, c.cache)
}

func (c *Client) VMs() api.VMs {
	return vms.New(c, c.cache)
}
