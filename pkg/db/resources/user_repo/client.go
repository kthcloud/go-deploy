package user_repo

import (
	"time"

	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db"
	"github.com/kthcloud/go-deploy/pkg/db/resources/base_clients"
	"go.mongodb.org/mongo-driver/bson"
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
	client.AddExtraFilter(bson.D{{Key: "lastAuthenticatedAt", Value: bson.D{{Key: "$gt", Value: lastAuthenticatedAt}}}})

	return client
}

// WithApiKey filters the users to only those with the given API key.
func (client *Client) WithApiKey(apiKey string) *Client {
	client.AddExtraFilter(bson.D{{Key: "apiKeys", Value: bson.D{{Key: "$elemMatch",
		Value: bson.D{
			{Key: "key", Value: apiKey},
			{Key: "expiresAt", Value: bson.D{{Key: "$gt", Value: time.Now()}}},
		},
	}}}})

	return client
}
