package system_gpu_info_repo

import (
	"go-deploy/dto/v2/body"
	"go-deploy/pkg/db"
	"go-deploy/pkg/db/resources/base_clients"
	"go.mongodb.org/mongo-driver/bson"
)

// Client is the client for the system GPU info model.
// It is used to query the database for data polled at a certain timestamp.
type Client struct {
	base_clients.TimestampedResourceClient[body.TimestampedSystemGpuInfo]
}

// New returns a new event client.
func New(n ...int) *Client {
	maxDocs := 1
	if len(n) > 0 {
		maxDocs = n[0]
	}

	return &Client{
		TimestampedResourceClient: base_clients.TimestampedResourceClient[body.TimestampedSystemGpuInfo]{
			Collection:   db.DB.GetCollection("systemGpuInfo"),
			MaxDocuments: maxDocs,
		},
	}
}

// AddExtraFilter adds an extra custom filter to the client.
func (client *Client) AddExtraFilter(filter bson.D) *Client {
	client.AddExtraFilter(filter)

	return client
}
