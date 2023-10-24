package notification

import (
	"go-deploy/models"
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
		Collection: models.NotificationCollection,

		ResourceClient: resource.ResourceClient[Notification]{
			Collection:     models.NotificationCollection,
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

func (client *Client) RestrictToUser(restrictUserID *string) *Client {
	if restrictUserID != nil {
		client.ResourceClient.ExtraFilter = &bson.D{{"userId", *restrictUserID}}
	}

	client.RestrictUserID = restrictUserID

	return client
}
