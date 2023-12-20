package job

import (
	"go-deploy/models/db"
	"go-deploy/models/sys/base"
	"go-deploy/models/sys/base/resource"
	"go.mongodb.org/mongo-driver/bson"
	"time"
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

func (client *Client) ExcludeScheduled() *Client {
	client.AddExtraFilter(bson.D{{"runAfter", bson.D{{"$lte", time.Now()}}}})

	return client
}

func (client *Client) ExcludeIDs(ids ...string) *Client {
	client.AddExtraFilter(bson.D{{"id", bson.D{{"$nin", ids}}}})

	return client
}

func (client *Client) AddFilter(filter bson.D) *Client {
	client.AddExtraFilter(filter)

	return client
}

func (client *Client) IncludeTypes(types ...string) *Client {
	client.AddExtraFilter(bson.D{{"type", bson.D{{"$in", types}}}})

	return client
}

func (client *Client) IncludeStatus(status ...string) *Client {
	client.AddExtraFilter(bson.D{{"status", bson.D{{"$in", status}}}})

	return client
}

func (client *Client) ExcludeStatus(status ...string) *Client {
	client.AddExtraFilter(bson.D{{"status", bson.D{{"$nin", status}}}})

	return client
}

func (client *Client) FilterArgs(argName string, filter interface{}) *Client {
	client.AddExtraFilter(bson.D{{"args." + argName, filter}})

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
	client.AddExtraFilter(bson.D{{"userId", restrictUserID}})
	client.RestrictUserID = &restrictUserID

	return client
}
