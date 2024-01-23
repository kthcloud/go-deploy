package vm_service

import (
	"go-deploy/service/vm_service/client"
)

// Client is the client for the Deployment service.
// It is used as a wrapper around the Client.
type Client struct {
	Version string `bson:"version"`

	client.BaseClient[Client]
}

// New creates a new deployment service client.
func New(version string) *Client {
	c := &Client{
		BaseClient: client.NewBaseClient[Client](nil),
	}
	c.BaseClient.SetParent(c)
	c.Version = version
	return c
}
