package team

import (
	"go-deploy/models/db"
	"go-deploy/models/sys/base"
	"go-deploy/models/sys/base/resource"
)

type Client struct {
	resource.ResourceClient[Team]
}

func New() *Client {
	return &Client{
		ResourceClient: resource.ResourceClient[Team]{
			Collection:     db.DB.GetCollection("teams"),
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
