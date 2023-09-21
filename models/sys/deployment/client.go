package deployment

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
