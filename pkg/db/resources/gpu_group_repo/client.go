package gpu_group_repo

import (
	"go-deploy/models/model"
	"go-deploy/pkg/db"
	"go-deploy/pkg/db/resources/base_clients"
	"go.mongodb.org/mongo-driver/bson"
)

// Client is used to manage GPU groups in the database.
type Client struct {
	base_clients.ResourceClient[model.GpuGroup]
}

// New returns a new GPU group client.
func New() *Client {
	return &Client{
		ResourceClient: base_clients.ResourceClient[model.GpuGroup]{
			Collection: db.DB.GetCollection("gpuGroups"),
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

// WithZone adds a filter to the client to only include groups with the given zone.
func (client *Client) WithZone(zone string) *Client {
	client.ResourceClient.AddExtraFilter(bson.D{{"zone", zone}})

	return client
}
