package vms

import (
	"go-deploy/service/clients"
	"go-deploy/service/core"
	"go-deploy/service/v1/vms/client"
	"go-deploy/service/v1/vms/cs_service"
	"go-deploy/service/v1/vms/k8s_service"
)

// Client is the client for the Deployment service.
// It is used as a wrapper around the Client.
type Client struct {
	// V1 is a reference to the parent client.
	V1 clients.V1

	client.BaseClient[Client]
}

// New creates a new VM service client.
func New(v1 clients.V1, cache ...*core.Cache) *Client {
	var ca *core.Cache
	if len(cache) > 0 {
		ca = cache[0]
	} else {
		ca = core.NewCache()
	}

	c := &Client{V1: v1, BaseClient: client.NewBaseClient[Client](ca)}
	c.BaseClient.SetParent(c)
	return c
}

// CS returns the client for the CS service.
func (c *Client) CS() *cs_service.Client {
	return cs_service.New(c.Cache)
}

// K8s returns the client for the K8s service.
func (c *Client) K8s() *k8s_service.Client {
	return k8s_service.New(c.Cache)
}
