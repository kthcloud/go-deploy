package job_repo

import (
	"time"

	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db"
	"github.com/kthcloud/go-deploy/pkg/db/resources/base_clients"
	"go.mongodb.org/mongo-driver/bson"
)

// Client is used to manage jobs in the database.
type Client struct {
	base_clients.ResourceClient[model.Job]
	RestrictUserID *string
}

// New returns a new job client.
func New() *Client {
	return &Client{
		ResourceClient: base_clients.ResourceClient[model.Job]{
			Collection:     db.DB.GetCollection("jobs"),
			IncludeDeleted: false,
			Pagination:     nil,
		},
	}
}

// ExcludeScheduled adds a filter to the client to exclude scheduled jobs.
func (client *Client) ExcludeScheduled() *Client {
	client.AddExtraFilter(bson.D{{Key: "runAfter", Value: bson.D{{Key: "$lte", Value: time.Now()}}}})

	return client
}

// ExcludeIDs adds a filter to the client to exclude the given IDs.
func (client *Client) ExcludeIDs(ids ...string) *Client {
	client.AddExtraFilter(bson.D{{Key: "id", Value: bson.D{{Key: "$nin", Value: ids}}}})

	return client
}

// AddFilter adds a custom filter to the client.
func (client *Client) AddFilter(filter bson.D) *Client {
	client.AddExtraFilter(filter)

	return client
}

// IncludeTypes adds a filter to the client to only include jobs with the given types.
func (client *Client) IncludeTypes(types ...string) *Client {
	client.AddExtraFilter(bson.D{{Key: "type", Value: bson.D{{Key: "$in", Value: types}}}})

	return client
}

// IncludeStatus adds a filter to the client to only include jobs with the given status.
func (client *Client) IncludeStatus(status ...string) *Client {
	client.AddExtraFilter(bson.D{{Key: "status", Value: bson.D{{Key: "$in", Value: status}}}})

	return client
}

// ExcludeStatus adds a filter to the client to exclude the given status.
func (client *Client) ExcludeStatus(status ...string) *Client {
	client.AddExtraFilter(bson.D{{Key: "status", Value: bson.D{{Key: "$nin", Value: status}}}})

	return client
}

// ExcludeTypes adds a filter to the client to exclude the given types.
func (client *Client) ExcludeTypes(types ...string) *Client {
	client.AddExtraFilter(bson.D{{Key: "type", Value: bson.D{{Key: "$nin", Value: types}}}})

	return client
}

// FilterArgs adds a filter to the client to only include jobs with the given args.
func (client *Client) FilterArgs(argName string, filter interface{}) *Client {
	client.AddExtraFilter(bson.D{{Key: "args." + argName, Value: filter}})

	return client
}

// WithPagination adds pagination to the client.
func (client *Client) WithPagination(page, pageSize int) *Client {
	client.Pagination = &db.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	return client
}

// WithSort adds sorting to the client.
func (client *Client) WithSort(field string, order int) *Client {
	client.SortBy = &db.SortBy{
		Field: field,
		Order: order,
	}

	return client
}

// WithUserID adds a filter to the client to only include jobs with the given user ID.
func (client *Client) WithUserID(userID string) *Client {
	client.AddExtraFilter(bson.D{{Key: "userId", Value: userID}})
	client.RestrictUserID = &userID

	return client
}
