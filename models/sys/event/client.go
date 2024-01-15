package event

import (
	"go-deploy/models/db"
	"go-deploy/models/sys/base"
	"go-deploy/models/sys/base/resource"
	"go.mongodb.org/mongo-driver/bson"
)

// Client is the client for the event resource.
// It is used to query the database for events, and extends resource.ResourceClient.
type Client struct {
	resource.ResourceClient[Event]
}

// New returns a new event client.
func New() *Client {
	return &Client{
		ResourceClient: resource.ResourceClient[Event]{
			Collection:     db.DB.GetCollection("events"),
			IncludeDeleted: false,
			Pagination:     nil,
		},
	}
}

// WithPagination sets the pagination for the client.
func (client *Client) WithPagination(page, pageSize int) *Client {
	client.Pagination = &base.Pagination{
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
