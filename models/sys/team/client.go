package team

import (
	"go-deploy/models/db"
	"go-deploy/models/sys/base"
	"go-deploy/models/sys/base/resource"
	"go.mongodb.org/mongo-driver/bson"
	"time"
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

func (client *Client) WithPagination(page, pageSize int) *Client {
	client.ResourceClient.Pagination = &base.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	return client
}

func (client *Client) WithUserID(userID string) *Client {
	client.AddExtraFilter(bson.D{{"$or", bson.A{
		bson.D{{"ownerId", userID}},
		bson.D{{"memberMap." + userID, bson.D{{"$exists", true}}}},
	}}})

	return client
}

func (client *Client) WithOwnerID(ownerID string) *Client {
	client.AddExtraFilter(bson.D{{"ownerId", ownerID}})

	return client
}

func (client *Client) WithResourceID(resourceID string) *Client {
	client.AddExtraFilter(bson.D{{"resourceMap." + resourceID, bson.D{{"$exists", true}}}})

	return client
}

func (client *Client) WithNameRegex(name string) *Client {
	filter := bson.D{{"name", bson.D{{"$regex", name}}}}
	client.ResourceClient.AddExtraFilter(filter)

	return client
}

func (client *Client) OlderThan(timestamp time.Time) *Client {
	filter := bson.D{{"createdAt", bson.D{{"$lt", timestamp}}}}
	client.ResourceClient.AddExtraFilter(filter)
	
	return client
}
