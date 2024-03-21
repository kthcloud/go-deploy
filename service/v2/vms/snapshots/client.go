package snapshots

import (
	"go-deploy/service/clients"
	"go-deploy/service/core"
	"go-deploy/service/v1/vms/client"
)

type Client struct {
	V1 clients.V1
	V2 clients.V2

	client.BaseClient[Client]
}

func New(v1 clients.V1, v2 clients.V2, cache ...*core.Cache) *Client {
	var ca *core.Cache
	if len(cache) > 0 {
		ca = cache[0]
	} else {
		ca = core.NewCache()
	}

	c := &Client{V1: v1, V2: v2, BaseClient: client.NewBaseClient[Client](ca)}
	c.BaseClient.SetParent(c)
	return c
}
