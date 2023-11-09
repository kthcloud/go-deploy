package team

import (
	"go-deploy/models/db"
	"go-deploy/models/sys/base"
	"go-deploy/models/sys/base/resource"
	"go.mongodb.org/mongo-driver/bson"
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

func (client *Client) AddUserID(userID string) *Client {
	client.AddExtraFilter(bson.D{{"$or", bson.A{
		bson.D{{"ownerId", userID}},
		bson.D{{"memberMap." + userID, bson.D{{"$exists", true}}}},
	}}})

	return client
}

func (client *Client) AddResourceID(resourceID string) *Client {
	client.AddExtraFilter(bson.D{{"resourceMap." + resourceID, bson.D{{"$exists", true}}}})

	return client
}
