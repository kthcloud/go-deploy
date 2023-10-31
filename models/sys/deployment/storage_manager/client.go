package storage_manager

import (
	"go-deploy/models/db"
	"go-deploy/models/sys/base"
	"go-deploy/models/sys/base/activityResource"
	"go-deploy/models/sys/base/resource"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Client struct {
	Collection      *mongo.Collection
	RestrictOwnerID *string

	activityResource.ActivityResourceClient[StorageManager]
	resource.ResourceClient[StorageManager]
}

func New() *Client {
	return &Client{
		Collection: db.DB.GetCollection("storageManagers"),

		ActivityResourceClient: activityResource.ActivityResourceClient[StorageManager]{
			Collection: db.DB.GetCollection("storageManagers"),
		},
		ResourceClient: resource.ResourceClient[StorageManager]{
			Collection:     db.DB.GetCollection("storageManagers"),
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

func (client *Client) RestrictToOwner(ownerID string) *Client {
	client.ResourceClient.ExtraFilter = append(client.ResourceClient.ExtraFilter, bson.E{Key: "ownerId", Value: ownerID})
	client.ActivityResourceClient.ExtraFilter = client.ResourceClient.ExtraFilter
	client.RestrictOwnerID = &ownerID

	return client
}
