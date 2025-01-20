package vm_port_repo

import (
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db"
	"github.com/kthcloud/go-deploy/pkg/db/resources/base_clients"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Client is the client for VM ports.
type Client struct {
	Collection *mongo.Collection

	base_clients.ResourceClient[model.VmPort]
}

// New creates a new VM port client.
func New() *Client {
	return &Client{
		Collection: db.DB.GetCollection("vmPorts"),

		ResourceClient: base_clients.ResourceClient[model.VmPort]{
			Collection:     db.DB.GetCollection("vmPorts"),
			IncludeDeleted: false,
		},
	}
}

// WithPagination sets the pagination for the client.
func (client *Client) WithPagination(page, pageSize int) *Client {
	client.ResourceClient.Pagination = &db.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	return client
}

// WithZone adds a filter to the client to only return VM ports in the given zone.
func (client *Client) WithZone(zone string) *Client {
	client.ResourceClient.AddExtraFilter(bson.D{{Key: "zone", Value: zone}})

	return client
}

// ExcludePortRange adds a filter to the client to exclude VM ports in the given range.
func (client *Client) ExcludePortRange(start, end int) *Client {
	filter := bson.D{{Key: "$or", Value: bson.A{
		bson.D{{Key: "publicPort", Value: bson.D{{Key: "$lt", Value: start}}}},
		bson.D{{Key: "publicPort", Value: bson.D{{Key: "$gte", Value: end}}}},
	}}}

	client.ResourceClient.AddExtraFilter(filter)

	return client
}

// IncludePortRange adds a filter to the client to only return VM ports in the given range.
func (client *Client) IncludePortRange(start, end int) *Client {
	client.ResourceClient.AddExtraFilter(bson.D{{Key: "publicPort", Value: bson.D{{Key: "$gte", Value: start}}}})
	client.ResourceClient.AddExtraFilter(bson.D{{Key: "publicPort", Value: bson.D{{Key: "$lt", Value: end}}}})

	return client
}

// WithVmID adds a filter to the client to only return VM ports for the given VM.
func (client *Client) WithVmID(vmID string) *Client {
	client.ResourceClient.AddExtraFilter(bson.D{{Key: "lease.vmId", Value: vmID}})

	return client
}
