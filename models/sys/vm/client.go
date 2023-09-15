package vm

import (
	"go-deploy/models"
	"go-deploy/models/sys/base/activityResource"
	"go-deploy/models/sys/base/resource"
	"go.mongodb.org/mongo-driver/mongo"
)

type Client struct {
	Collection *mongo.Collection

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
			Collection: models.VmCollection,
		},
	}
}

func NewWithDeleted() *Client {
	return &Client{
		Collection: models.VmCollection,

		ActivityResourceClient: activityResource.ActivityResourceClient[VM]{
			Collection: models.VmCollection,
		},
		ResourceClient: resource.ResourceClient[VM]{
			Collection:     models.VmCollection,
			IncludeDeleted: true,
		},
	}
}
