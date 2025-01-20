package api_key

import (
	"github.com/kthcloud/go-deploy/service/clients"
	"github.com/kthcloud/go-deploy/service/core"
)

type Client struct {
	V2 clients.V2

	Cache *core.Cache
}

func New(v2 clients.V2, cache ...*core.Cache) *Client {
	var ca *core.Cache
	if len(cache) > 0 {
		ca = cache[0]
	} else {
		ca = core.NewCache()
	}

	c := &Client{V2: v2, Cache: ca}
	return c
}
