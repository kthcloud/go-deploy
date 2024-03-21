package user_data_repo

import (
	"go-deploy/models/model"
	"go-deploy/pkg/db"
	"go-deploy/pkg/db/resources/base_clients"
	"go.mongodb.org/mongo-driver/bson"
)

// Client is used to manage users in the database.
type Client struct {
	base_clients.ResourceClient[model.UserData]
}

// New returns a new user client.
func New() *Client {
	return &Client{
		ResourceClient: base_clients.ResourceClient[model.UserData]{
			Collection:     db.DB.GetCollection("userData"),
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

// WithUserID adds a filter to the client to only return user data for the given user.
func (client *Client) WithUserID(userID string) *Client {
	client.AddExtraFilter(bson.D{{"userId", userID}})

	return client
}
