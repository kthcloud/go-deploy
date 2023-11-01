package user

import (
	"go-deploy/models"
	"go-deploy/models/db"
	"go-deploy/models/sys/base"
	"go-deploy/models/sys/base/resource"
)

type Client struct {
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
	client.Search = &models.SearchParams{
		Query:  search,
		Fields: db.DB.CollectionDefinitionMap["users"].TextIndexFields,
	}

	return client
}

func (client *Client) AddPagination(page, pageSize int) *Client {
	client.ResourceClient.Pagination = &base.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	return client
}
