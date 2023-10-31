package vm

import (
	"go-deploy/models/db"
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
		Collection: db.DB.GetCollection("vms"),

		ActivityResourceClient: activityResource.ActivityResourceClient[VM]{
			Collection: db.DB.GetCollection("vms"),
		},
		ResourceClient: resource.ResourceClient[VM]{
			Collection:     db.DB.GetCollection("vms"),
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

func (client *Client) RestrictToUser(restrictUserID string) *Client {
	client.ResourceClient.ExtraFilter = append(client.ResourceClient.ExtraFilter, bson.E{Key: "ownerId", Value: restrictUserID})
	client.ActivityResourceClient.ExtraFilter = client.ResourceClient.ExtraFilter
	client.RestrictUserID = &restrictUserID

	return client
}
