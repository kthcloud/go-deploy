package user

import (
	"go-deploy/models"
	"go-deploy/models/db"
	"go-deploy/models/sys/base"
	"go-deploy/models/sys/base/resource"
)

// Client is used to manage users in the database.
type Client struct {
	resource.ResourceClient[User]
}

// New returns a new user client.
func New() *Client {
	return &Client{
		ResourceClient: resource.ResourceClient[User]{
			Collection:     db.DB.GetCollection("users"),
			IncludeDeleted: false,
		},
	}
}

// WithSearch searches the `users` collection for the given search string.
// It uses the text index on the `users` collection to search.
func (client *Client) WithSearch(search string) *Client {
	client.Search = &models.SearchParams{
		Query:  search,
		Fields: db.DB.CollectionDefinitionMap["users"].TextIndexFields,
	}

	return client
}

// WithPagination adds pagination to the client.
func (client *Client) WithPagination(page, pageSize int) *Client {
	client.ResourceClient.Pagination = &base.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	return client
}
