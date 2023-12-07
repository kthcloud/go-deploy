package deployment_service

import (
	"go-deploy/service/deployment_service/client"
)

type Client struct {
	client.Client[Client]
}

func New() *Client {
	c := &Client{}
	c.Client.SetParent(c)
	return c
}
