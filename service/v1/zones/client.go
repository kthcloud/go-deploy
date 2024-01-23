package zones

import (
	"go-deploy/service/clients"
	"go-deploy/service/core"
)

// Client is the client for the Zone service.
type Client struct {
	// V1 is a reference to the parent client.
	V1 clients.V1

	// Cache is used to cache the resources fetched inside the service.
	Cache *core.Cache
}

// New creates a new zone service client.
func New(v1 clients.V1, cache ...*core.Cache) *Client {
	var c *core.Cache
	if len(cache) > 0 {
		c = cache[0]
	} else {
		c = core.NewCache()
	}

	return &Client{
		V1:    v1,
		Cache: c,
	}
}
