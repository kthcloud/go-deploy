package sms

import (
	"go-deploy/service/clients"
	"go-deploy/service/core"
	"go-deploy/service/v2/sms/client"
	"go-deploy/service/v2/sms/k8s_service"
)

// Client is the client for the Deployment service.
// It is used as a wrapper around the BaseClient.
type Client struct {
	// V2 is a reference to the parent client.
	V2 clients.V2

	client.BaseClient[Client]
}

// New creates a new SM service client.
func New(v2 clients.V2, cache ...*core.Cache) *Client {
	var ca *core.Cache
	if len(cache) > 0 {
		ca = cache[0]
	} else {
		ca = core.NewCache()
	}

	c := &Client{V2: v2, BaseClient: client.NewBaseClient[Client](ca)}
	c.BaseClient.SetParent(c)
	return c
}

// K8s returns the client for the K8s service.
func (c *Client) K8s() *k8s_service.Client {
	return k8s_service.New(c.Cache)
}
