package gpu

import (
	"go-deploy/models/db"
	"go-deploy/models/sys/base"
	"go-deploy/models/sys/base/resource"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

// Client is used to manage GPUs in the database.
type Client struct {
	resource.ResourceClient[GPU]
}

// New returns a new GPU client.
func New() *Client {
	return &Client{
		ResourceClient: resource.ResourceClient[GPU]{
			Collection: db.DB.GetCollection("gpus"),
		},
	}
}

// WithPagination sets the pagination for the client.
func (client *Client) WithPagination(page, pageSize int) *Client {
	client.ResourceClient.Pagination = &base.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	return client
}

// WithExclusion adds a filter to the client to exclude the given hosts and GPUs.
func (client *Client) WithExclusion(excludedHosts []string, excludedGPUs []string) *Client {
	if excludedHosts == nil {
		excludedHosts = make([]string, 0)
	}

	if excludedGPUs == nil {
		excludedGPUs = make([]string, 0)
	}

	filter := bson.D{
		{"host", bson.M{"$nin": excludedHosts}},
		{"id", bson.M{"$nin": excludedGPUs}},
	}

	client.ResourceClient.AddExtraFilter(filter)

	return client
}

// OnlyAvailable adds a filter to the client to only include available GPUs.
// Available GPUs are GPUs that are not leased, or whose lease has expired.
func (client *Client) OnlyAvailable() *Client {
	filter := bson.D{
		{"$or", []interface{}{
			bson.M{"lease": bson.M{"$exists": false}},
			bson.M{"lease.vmId": ""},
			bson.M{"lease.end": bson.M{"$lte": time.Now()}},
		}},
	}

	client.ResourceClient.AddExtraFilter(filter)

	return client
}

// WithVM adds a filter to the client to only include GPUs attached to the given VM.
func (client *Client) WithVM(vmID string) *Client {
	client.ResourceClient.AddExtraFilter(bson.D{{"lease.vmId", vmID}})

	return client
}

// WithZone adds a filter to the client to only include GPUs in the given zone.
func (client *Client) WithZone(zone string) *Client {
	client.ResourceClient.AddExtraFilter(bson.D{{"zone", zone}})

	return client
}

// ExcludeIDs adds a filter to the client to exclude the given GPU IDs.
func (client *Client) ExcludeIDs(ids ...string) *Client {
	client.ResourceClient.AddExtraFilter(bson.D{{"id", bson.M{"$nin": ids}}})

	return client
}

// WithoutLease adds a filter to the client to only include GPUs that are not leased.
func (client *Client) WithoutLease() *Client {
	filter := bson.D{
		{"$or", []interface{}{
			bson.M{"lease": bson.M{"$exists": false}},
			bson.M{"lease.vmId": ""},
		}},
	}

	client.ResourceClient.AddExtraFilter(filter)

	return client
}

// WithStale adds a filter to the client to only include GPUs that are stale.
func (client *Client) WithStale() *Client {
	client.ResourceClient.AddExtraFilter(bson.D{{"stale", true}})

	return client
}

// WithGroupName adds a filter to the client to only include GPUs with the given group name.
func (client *Client) WithGroupName(groupName string) *Client {
	client.ResourceClient.AddExtraFilter(bson.D{{"groupName", groupName}})

	return client
}
