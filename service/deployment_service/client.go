package deployment_service

import (
	"go-deploy/service/deployment_service/client"
)

type Client struct {
	client.BaseClient[Client]
}

func New() *Client {
	c := &Client{}
	c.BaseClient.SetParent(c)
	return c
}
