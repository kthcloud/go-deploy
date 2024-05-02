package api_key

import (
	"go-deploy/service/clients"
	"go-deploy/service/core"
)

type Client struct {
	V1 clients.V1
	V2 clients.V2

	Cache *core.Cache
}

func New(v1 clients.V1, v2 clients.V2, cache ...*core.Cache) *Client {
	var ca *core.Cache
	if len(cache) > 0 {
		ca = cache[0]
	} else {
		ca = core.NewCache()
	}

	c := &Client{V1: v1, V2: v2, Cache: ca}
	return c
}
