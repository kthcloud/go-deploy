package tmp_zone_repo

import (
	"go-deploy/models/model"
	"go-deploy/pkg/db"
	"go-deploy/pkg/db/resources/base_clients"
	"go.mongodb.org/mongo-driver/bson"
)

// Client is the client for the zone model.
// It is used to query the database for zones, and extends model.ResourceClient.
type Client struct {
	base_clients.ResourceClient[model.Zone]
}

// New returns a new zone client.
func New() *Client {
	return &Client{
		ResourceClient: base_clients.ResourceClient[model.Zone]{
			Collection:     db.DB.GetCollection("zones"),
			IncludeDeleted: false,
			Pagination:     nil,
		},
	}
}

// WithPagination sets the pagination for the client.
func (client *Client) WithPagination(page, pageSize int) *Client {
	client.Pagination = &db.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	return client
}

// AddExtraFilter adds an extra custom filter to the client.
func (client *Client) AddExtraFilter(filter bson.D) *Client {
	client.AddExtraFilter(filter)

	return client
}
