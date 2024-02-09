package user_data

import (
	"go-deploy/models/db"
	"go-deploy/models/sys/base"
	"go-deploy/models/sys/base/resource"
	"go.mongodb.org/mongo-driver/bson"
)

// Client is used to manage users in the database.
type Client struct {
	resource.ResourceClient[UserData]
}

// New returns a new user client.
func New() *Client {
	return &Client{
		ResourceClient: resource.ResourceClient[UserData]{
			Collection:     db.DB.GetCollection("userData"),
			IncludeDeleted: false,
		},
	}
}

// WithPagination adds pagination to the client.
func (client *Client) WithPagination(page, pageSize int) *Client {
	client.ResourceClient.Pagination = &base.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	return client
}

// WithUserID adds a filter to the client to only return user data for the given user.
func (client *Client) WithUserID(userID string) *Client {
	client.AddExtraFilter(bson.D{{"userId", userID}})

	return client
}
