package user

import (
	"go-deploy/models/db"
	"go-deploy/models/sys/base"
	"go-deploy/models/sys/base/resource"
)

type Client struct {
	Search *string

	resource.ResourceClient[User]
}

func New() *Client {
	return &Client{
		ResourceClient: resource.ResourceClient[User]{
			Collection:     db.DB.GetCollection("users"),
			IncludeDeleted: false,
		},
	}
}

func (client *Client) AddSearch(search string) *Client {
	client.Search = &search

	return client
}

func (client *Client) AddPagination(page, pageSize int) *Client {
	client.ResourceClient.Pagination = &base.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	return client
}
