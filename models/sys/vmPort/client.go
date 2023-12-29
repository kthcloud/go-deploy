package vmPort

import (
	"go-deploy/models/db"
	"go-deploy/models/sys/base"
	"go-deploy/models/sys/base/resource"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Client struct {
	Collection *mongo.Collection

	resource.ResourceClient[VmPort]
}

func New() *Client {
	return &Client{
		Collection: db.DB.GetCollection("vmPorts"),

		ResourceClient: resource.ResourceClient[VmPort]{
			Collection:     db.DB.GetCollection("vmPorts"),
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

func (client *Client) WithZone(zone string) *Client {
	client.ResourceClient.AddExtraFilter(bson.D{{"zone", zone}})

	return client
}

func (client *Client) ExcludePortRange(start, end int) *Client {
	filter := bson.D{{"$or", bson.A{
		bson.D{{"publicPort", bson.D{{"$lt", start}}}},
		bson.D{{"publicPort", bson.D{{"$gte", end}}}},
	}}}

	client.ResourceClient.AddExtraFilter(filter)

	return client
}

func (client *Client) IncludePortRange(start, end int) *Client {
	client.ResourceClient.AddExtraFilter(bson.D{{"publicPort", bson.D{{"$gte", start}}}})
	client.ResourceClient.AddExtraFilter(bson.D{{"publicPort", bson.D{{"$lt", end}}}})

	return client
}
