package user_repo

import (
	"go-deploy/models/model"
	"go-deploy/pkg/db"
	"go-deploy/pkg/db/resources/base_clients"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

// Client is used to manage users in the database.
type Client struct {
	base_clients.ResourceClient[model.User]
}

// New returns a new user client.
func New() *Client {
	return &Client{
		ResourceClient: base_clients.ResourceClient[model.User]{
			Collection:     db.DB.GetCollection("users"),
			IncludeDeleted: false,
		},
	}
}

// WithSearch searches the `users` collection for the given search string.
// It uses the text index on the `users` collection to search.
func (client *Client) WithSearch(search string) *Client {
	client.Search = &db.SearchParams{
		Query:  search,
		Fields: db.DB.CollectionDefinitionMap["users"].TextIndexFields,
	}

	return client
}

// WithPagination adds pagination to the client.
func (client *Client) WithPagination(page, pageSize int) *Client {
	client.ResourceClient.Pagination = &db.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	return client
}

// LastAuthenticatedAfter filters the users to only those who have authenticated after the given time.
func (client *Client) LastAuthenticatedAfter(lastAuthenticatedAt time.Time) *Client {
	client.AddExtraFilter(bson.D{{"lastAuthenticatedAt", bson.D{{"$gt", lastAuthenticatedAt}}}})

	return client
}
