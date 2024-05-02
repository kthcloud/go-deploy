package resource_migrations

import (
	"go-deploy/service/clients"
	"go-deploy/service/core"
)

type Client struct {
	V1 clients.V1
	V2 clients.V2

	// Cache is used to cache the resources fetched inside the service.
	Cache *core.Cache
}

// New creates a new resource migration service client.
func New(v1 clients.V1, v2 clients.V2, cache ...*core.Cache) *Client {
	var c *core.Cache
	if len(cache) > 0 {
		c = cache[0]
	} else {
		c = core.NewCache()
	}

	return &Client{
		V1:    v1,
		V2:    v2,
		Cache: c,
	}
}
