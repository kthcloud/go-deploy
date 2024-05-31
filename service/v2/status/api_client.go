package status

import (
	"go-deploy/service/clients"
	"go-deploy/service/core"
)

// Client is the client for the Status service.
type Client struct {
	// V2 is a reference to the parent client.
	V2 clients.V2

	// Cache is used to cache the resources fetched inside the service.
	Cache *core.Cache
}

// New creates a new Status service client.
func New(v2 clients.V2, cache ...*core.Cache) *Client {
	var c *core.Cache
	if len(cache) > 0 {
		c = cache[0]
	} else {
		c = core.NewCache()
	}

	return &Client{
		V2:    v2,
		Cache: c,
	}
}
