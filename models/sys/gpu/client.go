package gpu

import (
	"go-deploy/models/db"
	"go-deploy/models/sys/base"
	"go-deploy/models/sys/base/resource"
	"go.mongodb.org/mongo-driver/bson"
	"time"
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

func (client *Client) WithPagination(page, pageSize int) *Client {
	client.ResourceClient.Pagination = &base.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	return client
}

func (client *Client) WithExclusion(excludedHosts []string, excludedGPUs []string) *Client {
	if excludedHosts == nil {
		excludedHosts = make([]string, 0)
	}

	if excludedGPUs == nil {
		excludedGPUs = make([]string, 0)
	}

	client.ExcludedHosts = excludedHosts
	client.ExcludedGPUs = excludedGPUs

	return client
}

func (client *Client) OnlyAvailable() *Client {
	filter := bson.D{
		{"$or", []interface{}{
			bson.M{"lease": bson.M{"$exists": false}},
			bson.M{"lease.vmId": ""},
			bson.M{"lease.end": bson.M{"$lte": time.Now()}},
		}},
		{"host", bson.M{"$nin": client.ExcludedHosts}},
		{"id", bson.M{"$nin": client.ExcludedGPUs}},
	}

	client.ResourceClient.AddExtraFilter(filter)

	return client
}

func (client *Client) WithVM(vmID string) *Client {
	client.ResourceClient.AddExtraFilter(bson.D{{"lease.vmId", vmID}})

	return client
}
