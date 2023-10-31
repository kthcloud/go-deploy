package job

import (
	"go-deploy/models/db"
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
			Collection:     db.DB.GetCollection("jobs"),
			IncludeDeleted: false,
			Pagination:     nil,
		},
	}
}

func (client *Client) AddFilter(filter bson.D) *Client {
	client.ExtraFilter = filter

	return client
}

func (client *Client) AddPagination(page, pageSize int) *Client {
	client.Pagination = &base.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	return client
}

func (client *Client) RestrictToUser(restrictUserID string) *Client {
	client.ExtraFilter = append(client.ExtraFilter, bson.E{Key: "userId", Value: restrictUserID})
	client.RestrictUserID = &restrictUserID

	return client
}
