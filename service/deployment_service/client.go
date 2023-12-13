package deployment_service

import (
	"go-deploy/service/deployment_service/client"
)

// Client is the client for the Deployment service.
// It is used as a wrapper around the BaseClient.
type Client struct {
	client.BaseClient[Client]
}

// New creates a new deployment service client.
func New() *Client {
	c := &Client{
		BaseClient: client.NewBaseClient[Client](nil),
	}
	c.BaseClient.SetParent(c)
	return c
}
