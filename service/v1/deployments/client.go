package deployments

import (
	"go-deploy/service/clients"
	"go-deploy/service/core"
	"go-deploy/service/v1/deployments/client"
	"go-deploy/service/v1/deployments/harbor_service"
	"go-deploy/service/v1/deployments/k8s_service"
)

// Client is the client for the Deployment service.
// It is used as a wrapper around the BaseClient.
type Client struct {
	// V1 is a reference to the parent client.
	V1 clients.V1

	client.BaseClient[Client]
}

// New creates a new deployment service client.
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

// Harbor returns the client for the Harbor service.
func (c *Client) Harbor() *harbor_service.Client {
	return harbor_service.New(c.Cache)
}

// K8s returns the client for the K8s service.
func (c *Client) K8s() *k8s_service.Client {
	return k8s_service.New(c.Cache)
}
