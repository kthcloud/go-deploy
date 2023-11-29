package gpu

import (
	"go-deploy/models/db"
	"go-deploy/models/sys/base/resource"
	"go.mongodb.org/mongo-driver/bson"
)

type Client struct {
	ExcludedHosts []string
	ExcludedGPUs  []string

	resource.ResourceClient[GPU]
}

func New() *Client {
	return &Client{
		ExcludedHosts: make([]string, 0),
		ExcludedGPUs:  make([]string, 0),
		ResourceClient: resource.ResourceClient[GPU]{
			Collection: db.DB.GetCollection("gpus"),
		},
	}
}

func NewWithExclusion(excludedHosts []string, excludedGPUs []string) *Client {
	if excludedHosts == nil {
		excludedHosts = make([]string, 0)
	}

	if excludedGPUs == nil {
		excludedGPUs = make([]string, 0)
	}

	return &Client{
		ExcludedHosts: excludedHosts,
		ExcludedGPUs:  excludedGPUs,
		ResourceClient: resource.ResourceClient[GPU]{
			Collection: db.DB.GetCollection("gpus"),
		},
	}
}

func (client *Client) WithVM(vmID string) *Client {
	client.ResourceClient.AddExtraFilter(bson.D{{"lease.vmId", vmID}})

	return client
}
