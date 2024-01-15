package job

import (
	"go-deploy/models/db"
	"go-deploy/models/sys/base"
	"go-deploy/models/sys/base/resource"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

// Client is used to manage jobs in the database.
type Client struct {
	resource.ResourceClient[Job]
	RestrictUserID *string
}

// New returns a new job client.
func New() *Client {
	return &Client{
		ResourceClient: resource.ResourceClient[Job]{
			Collection:     db.DB.GetCollection("jobs"),
			IncludeDeleted: false,
			Pagination:     nil,
		},
	}
}

// ExcludeScheduled adds a filter to the client to exclude scheduled jobs.
func (client *Client) ExcludeScheduled() *Client {
	client.AddExtraFilter(bson.D{{"runAfter", bson.D{{"$lte", time.Now()}}}})

	return client
}

// ExcludeIDs adds a filter to the client to exclude the given IDs.
func (client *Client) ExcludeIDs(ids ...string) *Client {
	client.AddExtraFilter(bson.D{{"id", bson.D{{"$nin", ids}}}})

	return client
}

// AddFilter adds a custom filter to the client.
func (client *Client) AddFilter(filter bson.D) *Client {
	client.AddExtraFilter(filter)

	return client
}

// IncludeTypes adds a filter to the client to only include jobs with the given types.
func (client *Client) IncludeTypes(types ...string) *Client {
	client.AddExtraFilter(bson.D{{"type", bson.D{{"$in", types}}}})

	return client
}

// IncludeStatus adds a filter to the client to only include jobs with the given status.
func (client *Client) IncludeStatus(status ...string) *Client {
	client.AddExtraFilter(bson.D{{"status", bson.D{{"$in", status}}}})

	return client
}

// ExcludeStatus adds a filter to the client to exclude the given status.
func (client *Client) ExcludeStatus(status ...string) *Client {
	client.AddExtraFilter(bson.D{{"status", bson.D{{"$nin", status}}}})

	return client
}

// ExcludeTypes adds a filter to the client to exclude the given types.
func (client *Client) ExcludeTypes(types ...string) *Client {
	client.AddExtraFilter(bson.D{{"type", bson.D{{"$nin", types}}}})

	return client
}

// FilterArgs adds a filter to the client to only include jobs with the given args.
func (client *Client) FilterArgs(argName string, filter interface{}) *Client {
	client.AddExtraFilter(bson.D{{"args." + argName, filter}})

	return client
}

// WithPagination adds pagination to the client.
func (client *Client) WithPagination(page, pageSize int) *Client {
	client.Pagination = &base.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	return client
}

// RestrictToUser adds a filter to the client to only include jobs with the given user ID.
func (client *Client) RestrictToUser(restrictUserID string) *Client {
	client.AddExtraFilter(bson.D{{"userId", restrictUserID}})
	client.RestrictUserID = &restrictUserID

	return client
}
