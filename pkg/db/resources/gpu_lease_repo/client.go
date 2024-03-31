package gpu_lease_repo

import (
	"go-deploy/models/model"
	"go-deploy/pkg/db"
	"go-deploy/pkg/db/resources/base_clients"
	"go.mongodb.org/mongo-driver/bson"
)

// Client is used to manage GPU leases in the database.
type Client struct {
	base_clients.ResourceClient[model.GpuLease]
}

// New returns a new GPU lease client.
func New() *Client {
	return &Client{
		ResourceClient: base_clients.ResourceClient[model.GpuLease]{
			Collection: db.DB.GetCollection("gpuLeases"),
		},
	}
}

// WithPagination sets the pagination for the client.
func (client *Client) WithPagination(page, pageSize int) *Client {
	client.ResourceClient.Pagination = &db.Pagination{
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

// WithVmID adds a filter to the client to only include leases with the given VM ID.
func (client *Client) WithVmID(vmID string) *Client {
	client.ResourceClient.AddExtraFilter(bson.D{{"vmId", vmID}})

	return client
}

// WithUserID adds a filter to the client to only include leases with the given user ID.
func (client *Client) WithUserID(userID string) *Client {
	client.ResourceClient.AddExtraFilter(bson.D{{"userId", userID}})

	return client
}

// OnlyActive adds a filter to the client to only include active leases.
func (client *Client) OnlyActive() *Client {
	// An active lease is one that has a expiresAt field set
	filter := bson.D{{"expiresAt", bson.D{{"$exists", true}}}}
	client.ResourceClient.AddExtraFilter(filter)

	return client
}
