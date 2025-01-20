package vm_repo

import (
	"time"

	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db"
	"github.com/kthcloud/go-deploy/pkg/db/resources/base_clients"
	"github.com/kthcloud/go-deploy/service/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Client is used to manage VMs in the database.
type Client struct {
	Collection     *mongo.Collection
	RestrictUserID *string
	Version        *string

	base_clients.ActivityResourceClient[model.VM]
	base_clients.ResourceClient[model.VM]
}

// New returns a new VM client.
func New(version ...string) *Client {
	c := &Client{
		Collection: db.DB.GetCollection("vms"),

		ActivityResourceClient: base_clients.ActivityResourceClient[model.VM]{
			Collection: db.DB.GetCollection("vms"),
		},
		ResourceClient: base_clients.ResourceClient[model.VM]{
			Collection:     db.DB.GetCollection("vms"),
			IncludeDeleted: false,
		},
	}

	if ver := utils.GetFirstOrDefault(version); ver != "" {
		c.WithVersion(ver)
	}

	return c
}

// WithVersion adds a filter to the client to only return VMs with the given version.
func (client *Client) WithVersion(version string) *Client {
	client.ResourceClient.AddExtraFilter(bson.D{{Key: "version", Value: version}})
	client.ActivityResourceClient.ExtraFilter = client.ResourceClient.ExtraFilter

	return client
}

// WithPagination adds pagination to the client.
func (client *Client) WithPagination(page, pageSize int) *Client {
	client.ResourceClient.Pagination = &db.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	client.ActivityResourceClient.Pagination = &db.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	return client
}

// IncludeDeletedResources makes the client include deleted VMs.
func (client *Client) IncludeDeletedResources() *Client {
	client.IncludeDeleted = true

	return client
}

// WithOwner adds a filter to the client to only return VMs owned by the given ownerID.
func (client *Client) WithOwner(ownerID string) *Client {
	client.ResourceClient.AddExtraFilter(bson.D{{Key: "ownerId", Value: ownerID}})
	client.ActivityResourceClient.ExtraFilter = client.ResourceClient.ExtraFilter
	client.RestrictUserID = &ownerID

	return client
}

// WithActivities adds a filter to the client to only return VMs that have the given activities.
func (client *Client) WithActivities(activities ...string) *Client {
	orFilter := bson.A{}

	for _, activity := range activities {
		orFilter = append(orFilter, bson.M{
			"activities." + activity: bson.M{
				"$exists": true,
			},
		})
	}

	filter := bson.D{{
		Key: "$or", Value: orFilter,
	}}

	client.ResourceClient.AddExtraFilter(filter)
	client.ActivityResourceClient.AddExtraFilter(filter)

	return client
}

// WithNoActivities adds a filter to the client to only return VMs that have no activities.
func (client *Client) WithNoActivities() *Client {
	filter := bson.D{{
		Key: "activities", Value: bson.M{
			"$gte": bson.M{},
		},
	}}

	client.ResourceClient.AddExtraFilter(filter)
	client.ActivityResourceClient.AddExtraFilter(filter)

	return client
}

// WithIDs adds a filter to the client to only return VMs with the given IDs.
func (client *Client) WithIDs(ids ...string) *Client {
	filter := bson.D{{Key: "id", Value: bson.D{{Key: "$in", Value: ids}}}}

	client.ResourceClient.AddExtraFilter(filter)
	client.ActivityResourceClient.AddExtraFilter(filter)

	return client
}

// WithCustomFilter adds a custom filter to the client.
func (client *Client) WithCustomFilter(filter bson.D) *Client {
	client.ResourceClient.AddExtraFilter(filter)
	client.ActivityResourceClient.AddExtraFilter(filter)

	return client
}

// WithNameRegex adds a filter to the client to only return VMs with names matching the given regex.
func (client *Client) WithNameRegex(name string) *Client {
	filter := bson.D{{Key: "name", Value: bson.D{{Key: "$regex", Value: name}}}}

	client.ResourceClient.AddExtraFilter(filter)
	client.ActivityResourceClient.AddExtraFilter(filter)

	return client
}

// WithZone adds a filter to the client to only return VMs in the given zone.
func (client *Client) WithZone(zone ...string) *Client {
	filter := bson.D{{Key: "zone", Value: bson.D{{Key: "$in", Value: zone}}}}

	client.ResourceClient.AddExtraFilter(filter)
	client.ActivityResourceClient.AddExtraFilter(filter)

	return client
}

// OlderThan adds a filter to the client to only return VMs created before the given timestamp.
func (client *Client) OlderThan(timestamp time.Time) *Client {
	filter := bson.D{{Key: "createdAt", Value: bson.D{{Key: "$lt", Value: timestamp}}}}

	client.ResourceClient.AddExtraFilter(filter)
	client.ActivityResourceClient.AddExtraFilter(filter)

	return client
}

// LastAccessedBefore adds a filter to the client to only return VMs that were last accessed before the given timestamp.
func (client *Client) LastAccessedBefore(timestamp time.Time) *Client {
	filter := bson.D{{Key: "accessedAt", Value: bson.D{{Key: "$lt", Value: timestamp}}}}

	client.ResourceClient.AddExtraFilter(filter)
	client.ActivityResourceClient.AddExtraFilter(filter)

	return client
}
