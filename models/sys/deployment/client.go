package deployment

import (
	"go-deploy/models"
	"go-deploy/models/sys/base/activityResource"
	"go-deploy/models/sys/base/resource"
	"go.mongodb.org/mongo-driver/mongo"
)

type Client struct {
	Collection *mongo.Collection

	activityResource.ActivityResourceClient[Deployment]
	resource.ResourceClient[Deployment]
}

func New() *Client {
	return &Client{
		Collection: models.DeploymentCollection,

		ActivityResourceClient: activityResource.ActivityResourceClient[Deployment]{
			Collection: models.DeploymentCollection,
		},
		ResourceClient: resource.ResourceClient[Deployment]{
			Collection:     models.DeploymentCollection,
			IncludeDeleted: false,
		},
	}
}

func NewWithDeleted() *Client {
	return &Client{
		Collection: models.DeploymentCollection,

		ActivityResourceClient: activityResource.ActivityResourceClient[Deployment]{
			Collection: models.DeploymentCollection,
		},
		ResourceClient: resource.ResourceClient[Deployment]{
			Collection:     models.DeploymentCollection,
			IncludeDeleted: true,
		},
	}
}
