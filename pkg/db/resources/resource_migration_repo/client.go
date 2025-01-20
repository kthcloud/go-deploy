package resource_migration_repo

import (
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db"
	"github.com/kthcloud/go-deploy/pkg/db/resources/base_clients"
	"go.mongodb.org/mongo-driver/bson"
)

type Client struct {
	base_clients.ResourceClient[model.ResourceMigration]
}

func New() *Client {
	return &Client{
		ResourceClient: base_clients.ResourceClient[model.ResourceMigration]{
			Collection: db.DB.GetCollection("resourceMigrations"),
		},
	}
}

// WithPagination sets the pagination for the client.
func (c *Client) WithPagination(page, pageSize int) *Client {
	c.ResourceClient.Pagination = &db.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	return c
}

// WithType adds a filter to the client to only include resource migrations with the given type.
func (c *Client) WithType(migrationType string) *Client {
	c.ResourceClient.AddExtraFilter(bson.D{{Key: "type", Value: migrationType}})

	return c
}

// WithResourceType adds a filter to the client to only include resource migrations with the given resource type.
func (c *Client) WithResourceType(resourceType string) *Client {
	c.ResourceClient.AddExtraFilter(bson.D{{Key: "resourceType", Value: resourceType}})

	return c
}

// WithResourceID adds a filter to the client to only include resource migrations with the given resource ID.
func (c *Client) WithResourceID(resourceID string) *Client {
	c.ResourceClient.AddExtraFilter(bson.D{{Key: "resourceId", Value: resourceID}})

	return c
}

// WithUserID adds a filter to the client to only include resource migrations with the given user ID.
func (c *Client) WithUserID(userID string) *Client {
	c.ResourceClient.AddExtraFilter(bson.D{{Key: "userId", Value: userID}})

	return c
}

// WithStatus adds a filter to the client to only include resource migrations with the given status.
func (c *Client) WithStatus(status string) *Client {
	c.ResourceClient.AddExtraFilter(bson.D{{Key: "status", Value: status}})

	return c
}

// WithCode adds a filter to the client to only include resource migrations with the given code.
func (c *Client) WithCode(code string) *Client {
	c.ResourceClient.AddExtraFilter(bson.D{{Key: "code", Value: code}})

	return c
}
