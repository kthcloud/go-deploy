package team

import (
	"go-deploy/models"
	"go-deploy/models/sys/base"
	"go-deploy/models/sys/base/resource"
)

type Client struct {
	resource.ResourceClient[Team]
}

func New() *Client {
	return &Client{
		ResourceClient: resource.ResourceClient[Team]{
			Collection:     models.UserCollection,
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
