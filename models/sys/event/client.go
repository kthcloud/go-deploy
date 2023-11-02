package event

import (
	"go-deploy/models/db"
	"go-deploy/models/sys/base"
	"go-deploy/models/sys/base/resource"
	"go.mongodb.org/mongo-driver/bson"
)

type Client struct {
	resource.ResourceClient[Event]
}

func New() *Client {
	return &Client{
		ResourceClient: resource.ResourceClient[Event]{
			Collection:     db.DB.GetCollection("events"),
			IncludeDeleted: false,
			Pagination:     nil,
		},
	}
}

func (client *Client) AddPagination(page, pageSize int) *Client {
	client.Pagination = &base.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	return client
}

func (client *Client) AddExtraFilter(filter bson.D) *Client {
	client.AddExtraFilter(filter)

	return client
}
