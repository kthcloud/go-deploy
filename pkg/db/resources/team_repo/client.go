package team_repo

import (
	"time"

	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db"
	"github.com/kthcloud/go-deploy/pkg/db/resources/base_clients"
	"go.mongodb.org/mongo-driver/bson"
)

// Client is used to manage teams in the database.
type Client struct {
	base_clients.ResourceClient[model.Team]
}

// New returns a new team client.
func New() *Client {
	return &Client{
		ResourceClient: base_clients.ResourceClient[model.Team]{
			Collection:     db.DB.GetCollection("teams"),
			IncludeDeleted: false,
		},
	}
}

// WithPagination adds pagination to the client.
func (client *Client) WithPagination(page, pageSize int) *Client {
	client.ResourceClient.Pagination = &db.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	return client
}

// WithUserID adds a filter to the client to only include teams with the given user ID.
func (client *Client) WithUserID(userID string) *Client {
	client.AddExtraFilter(bson.D{{Key: "$or", Value: bson.A{
		bson.D{{Key: "ownerId", Value: userID}},
		bson.D{{Key: "memberMap." + userID, Value: bson.D{{Key: "$exists", Value: true}}}},
	}}})

	return client
}

// WithOwnerID adds a filter to the client to only include teams with the given owner ID.
func (client *Client) WithOwnerID(ownerID string) *Client {
	client.AddExtraFilter(bson.D{{Key: "ownerId", Value: ownerID}})

	return client
}

// WithResourceID adds a filter to the client to only include teams with the given model ID.
func (client *Client) WithResourceID(resourceID string) *Client {
	client.AddExtraFilter(bson.D{{Key: "resourceMap." + resourceID, Value: bson.D{{Key: "$exists", Value: true}}}})

	return client
}

// WithNameRegex adds a filter to the client to only include teams with a name matching the given regex.
func (client *Client) WithNameRegex(name string) *Client {
	filter := bson.D{{Key: "name", Value: bson.D{{Key: "$regex", Value: name}}}}
	client.ResourceClient.AddExtraFilter(filter)

	return client
}

// OlderThan adds a filter to the client to only include teams older than the given timestamp.
func (client *Client) OlderThan(timestamp time.Time) *Client {
	filter := bson.D{{Key: "createdAt", Value: bson.D{{Key: "$lt", Value: timestamp}}}}
	client.ResourceClient.AddExtraFilter(filter)

	return client
}
