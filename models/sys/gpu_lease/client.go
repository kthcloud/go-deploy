package gpu_lease

import (
	"go-deploy/models/db"
	"go-deploy/models/sys/base"
	"go-deploy/models/sys/base/resource"
	"go.mongodb.org/mongo-driver/bson"
)

// Client is used to manage GPUs in the database.
type Client struct {
	resource.ResourceClient[GpuLease]
}

// New returns a new GPU client.
func New() *Client {
	return &Client{
		ResourceClient: resource.ResourceClient[GpuLease]{
			Collection: db.DB.GetCollection("gpuLeases"),
		},
	}
}

// WithPagination sets the pagination for the client.
func (client *Client) WithPagination(page, pageSize int) *Client {
	client.ResourceClient.Pagination = &base.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	return client
}

// WithGroupName adds a filter to the client to only include leases with the given group name.
func (client *Client) WithGroupName(groupName string) *Client {
	client.ResourceClient.AddExtraFilter(bson.D{{"groupName", groupName}})

	return client
}

// WithVM adds a filter to the client to only include leases with the given VM ID.
func (client *Client) WithVM(vmID string) *Client {
	client.ResourceClient.AddExtraFilter(bson.D{{"vmId", vmID}})

	return client
}

// OnlyActive adds a filter to the client to only include active leases.
func (client *Client) OnlyActive() *Client {
	// An active lease is one that has a expiresAt field set
	filter := bson.D{{"expiresAt", bson.D{{"$exists", true}}}}
	client.ResourceClient.AddExtraFilter(filter)

	return client
}
