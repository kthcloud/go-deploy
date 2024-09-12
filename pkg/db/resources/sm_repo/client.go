package sm_repo

import (
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db"
	"github.com/kthcloud/go-deploy/pkg/db/resources/base_clients"
	"go.mongodb.org/mongo-driver/bson"
)

// Client is used to manage storage managers in the database.
type Client struct {
	base_clients.ResourceClient[model.SM]
	base_clients.ActivityResourceClient[model.SM]
}

// New returns a new storage manager client.
func New() *Client {
	return &Client{
		ResourceClient: base_clients.ResourceClient[model.SM]{
			Collection:     db.DB.GetCollection("storageManagers"),
			IncludeDeleted: false,
		},
		ActivityResourceClient: base_clients.ActivityResourceClient[model.SM]{
			Collection: db.DB.GetCollection("storageManagers"),
		},
	}
}

// WithPagination adds pagination to the client.
func (client *Client) WithPagination(page, pageSize int) *Client {
	pagination := &db.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	client.ResourceClient.Pagination = pagination
	client.ActivityResourceClient.Pagination = pagination

	return client
}

// IncludeDeletedResources makes the client include deleted storage managers.
func (client *Client) IncludeDeletedResources() *Client {
	client.ResourceClient.IncludeDeleted = true

	return client
}

// WithOwnerID adds a filter to the client to only include storage managers with the given owner ID.
func (client *Client) WithOwnerID(ownerID string) *Client {
	filter := bson.D{{"ownerId", ownerID}}

	client.ResourceClient.AddExtraFilter(filter)
	client.ActivityResourceClient.AddExtraFilter(filter)

	return client
}

// WithActivities adds a filter to the client to only include storage managers that does
// at least one of the given activities.
func (client *Client) WithActivities(activities ...string) *Client {
	orFilter := bson.A{}

	for _, activity := range activities {
		orFilter = append(orFilter, bson.M{
			"activities." + activity: bson.M{
				"$exists": true,
			},
		})
	}

	filter := bson.D{{
		"$or", orFilter,
	}}

	client.ResourceClient.AddExtraFilter(filter)
	client.ActivityResourceClient.AddExtraFilter(filter)

	return client
}

// WithNoActivities adds a filter to the client to only include storage managers without any activities.
func (client *Client) WithNoActivities() *Client {
	filter := bson.D{{
		"activities", bson.M{
			"$gte": bson.M{},
		},
	}}

	client.ResourceClient.AddExtraFilter(filter)
	client.ActivityResourceClient.AddExtraFilter(filter)

	return client
}

// WithZone adds a filter to the client to only include storage managers with the given zone.
func (client *Client) WithZone(zone string) *Client {
	filter := bson.D{{"zone", zone}}

	client.ResourceClient.AddExtraFilter(filter)
	client.ActivityResourceClient.AddExtraFilter(filter)

	return client
}
