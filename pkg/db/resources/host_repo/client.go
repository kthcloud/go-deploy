package host_repo

import (
	"time"

	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db"
	"github.com/kthcloud/go-deploy/pkg/db/resources/base_clients"
	"go.mongodb.org/mongo-driver/bson"
)

// Client is the client for the host model.
// It is used to query the database for hosts, and extends model.ResourceClient.
type Client struct {
	base_clients.ResourceClient[model.Host]
}

// New returns a new host client.
func New() *Client {
	return &Client{
		ResourceClient: base_clients.ResourceClient[model.Host]{
			Collection:     db.DB.GetCollection("hosts"),
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

// Activated adds a filter to only return activated hosts
func (client *Client) Activated() *Client {
	filter := bson.D{{Key: "$or", Value: bson.A{
		bson.D{{Key: "deactivatedUntil", Value: bson.D{{Key: "$lt", Value: time.Now()}}}},
		bson.D{{Key: "deactivatedUntil", Value: bson.D{{Key: "$exists", Value: false}}}},
	}}}

	client.AddExtraFilter(filter)

	return client
}

// Schedulable adds a filter to only return hosts that are schedulable
func (client *Client) Schedulable() *Client {
	filter := bson.D{{Key: "schedulable", Value: true}}

	client.AddExtraFilter(filter)

	return client
}
