package v1

import (
	"go-deploy/service/core"
	"go-deploy/service/v1/api"
	"go-deploy/service/v1/deployments"
	"go-deploy/service/v1/discovery"
	"go-deploy/service/v1/events"
	"go-deploy/service/v1/jobs"
	"go-deploy/service/v1/notifications"
	"go-deploy/service/v1/sms"
	"go-deploy/service/v1/status"
	"go-deploy/service/v1/teams"
	"go-deploy/service/v1/user_data"
	"go-deploy/service/v1/users"
	"go-deploy/service/v1/vms"
	"go-deploy/service/v1/zones"
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
	return users.New(c, c.cache)
}

func (c *Client) UserData() api.UserData {
	return user_data.New(c, c.cache)
}

func (c *Client) VMs() api.VMs {
	return vms.New(c, c.cache)
}

func (c *Client) Zones() api.Zones {
	return zones.New(c, c.cache)
}
