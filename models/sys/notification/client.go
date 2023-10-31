package notification

import (
	"go-deploy/models/db"
	"go-deploy/models/sys/base"
	"go-deploy/models/sys/base/resource"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Client struct {
	Collection     *mongo.Collection
	RestrictUserID *string

	resource.ResourceClient[Notification]
}

func New() *Client {
	return &Client{
		Collection: db.DB.GetCollection("notifications"),

		ResourceClient: resource.ResourceClient[Notification]{
			Collection:     db.DB.GetCollection("notifications"),
			IncludeDeleted: false,
		},
	}
}

func (client *Client) AddPagination(page, pageSize int) *Client {
	client.ResourceClient.Pagination = &base.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	return client
}

func (client *Client) RestrictToUser(restrictUserID string) *Client {
	client.ResourceClient.ExtraFilter = append(client.ResourceClient.ExtraFilter, bson.E{Key: "ownerId", Value: restrictUserID})
	client.RestrictUserID = &restrictUserID

	return client
}
