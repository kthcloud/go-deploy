package job

import (
	"go-deploy/models"
	"go-deploy/models/sys/base"
	"go-deploy/models/sys/base/resource"
	"go.mongodb.org/mongo-driver/bson"
)

type Client struct {
	resource.ResourceClient[Job]
	RestrictUserID *string
}

func New() *Client {
	return &Client{
		ResourceClient: resource.ResourceClient[Job]{
			Collection:     models.JobCollection,
			IncludeDeleted: false,
			Pagination:     nil,
		},
	}
}

func (client *Client) AddFilter(filter bson.D) *Client {
	client.ExtraFilter = &filter

	return client
}

func (client *Client) AddPagination(page, pageSize int) *Client {
	client.Pagination = &base.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	return client
}

func (client *Client) AddRestrictedUser(restrictUserID *string) *Client {
	if restrictUserID != nil {
		client.ExtraFilter = &bson.D{{"userId", *restrictUserID}}
	}

	client.RestrictUserID = restrictUserID

	return client
}
