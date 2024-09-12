package team_repo

import (
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db"
	"github.com/kthcloud/go-deploy/pkg/db/resources/base_clients"
	"go.mongodb.org/mongo-driver/bson"
	"time"
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
	client.AddExtraFilter(bson.D{{"$or", bson.A{
		bson.D{{"ownerId", userID}},
		bson.D{{"memberMap." + userID, bson.D{{"$exists", true}}}},
	}}})

	return client
}

// WithOwnerID adds a filter to the client to only include teams with the given owner ID.
func (client *Client) WithOwnerID(ownerID string) *Client {
	client.AddExtraFilter(bson.D{{"ownerId", ownerID}})

	return client
}

// WithResourceID adds a filter to the client to only include teams with the given model ID.
func (client *Client) WithResourceID(resourceID string) *Client {
	client.AddExtraFilter(bson.D{{"resourceMap." + resourceID, bson.D{{"$exists", true}}}})

	return client
}

// WithNameRegex adds a filter to the client to only include teams with a name matching the given regex.
func (client *Client) WithNameRegex(name string) *Client {
	filter := bson.D{{"name", bson.D{{"$regex", name}}}}
	client.ResourceClient.AddExtraFilter(filter)

	return client
}

// OlderThan adds a filter to the client to only include teams older than the given timestamp.
func (client *Client) OlderThan(timestamp time.Time) *Client {
	filter := bson.D{{"createdAt", bson.D{{"$lt", timestamp}}}}
	client.ResourceClient.AddExtraFilter(filter)

	return client
}
