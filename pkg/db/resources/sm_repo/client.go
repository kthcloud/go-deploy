package sm_repo

import (
	"go-deploy/models/model"
	"go-deploy/pkg/db"
	"go-deploy/pkg/db/resources/base_clients"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Client is used to manage storage managers in the database.
type Client struct {
	Collection      *mongo.Collection
	RestrictOwnerID *string

	base_clients.ResourceClient[model.SM]
}

// New returns a new storage manager client.
func New() *Client {
	return &Client{
		Collection: db.DB.GetCollection("storageManagers"),

		ResourceClient: base_clients.ResourceClient[model.SM]{
			Collection:     db.DB.GetCollection("storageManagers"),
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

// IncludeDeletedResources makes the client include deleted storage managers.
func (client *Client) IncludeDeletedResources() *Client {
	client.IncludeDeleted = true

	return client
}

// RestrictToOwner adds a filter to the client to only include storage managers with the given owner ID.
func (client *Client) RestrictToOwner(ownerID string) *Client {
	client.ResourceClient.AddExtraFilter(bson.D{{"ownerId", ownerID}})
	client.RestrictOwnerID = &ownerID

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

	return client
}
