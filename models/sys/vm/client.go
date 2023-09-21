package vm

import (
	"go-deploy/models"
	"go-deploy/models/sys/base"
	"go-deploy/models/sys/base/activityResource"
	"go-deploy/models/sys/base/resource"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Client struct {
	Collection     *mongo.Collection
	RestrictUserID *string

	activityResource.ActivityResourceClient[VM]
	resource.ResourceClient[VM]
}

func New() *Client {
	return &Client{
		Collection: models.VmCollection,

		ActivityResourceClient: activityResource.ActivityResourceClient[VM]{
			Collection: models.VmCollection,
		},
		ResourceClient: resource.ResourceClient[VM]{
			Collection:     models.VmCollection,
			IncludeDeleted: false,
		},
	}
}

func (client *Client) AddPagination(page, pageSize int) *Client {
	client.ResourceClient.Pagination = &base.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	client.ActivityResourceClient.Pagination = &base.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	return client
}

func (client *Client) IncludeDeletedResources() *Client {
	client.IncludeDeleted = true

	return client
}

func (client *Client) RestrictToUser(restrictUserID *string) *Client {
	if restrictUserID != nil {
		client.ResourceClient.ExtraFilter = &bson.D{{"ownerId", *restrictUserID}}
		client.ActivityResourceClient.ExtraFilter = &bson.D{{"ownerId", *restrictUserID}}
	}

	client.RestrictUserID = restrictUserID

	return client
}
