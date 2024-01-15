package worker_status

import (
	"go-deploy/models/db"
	"go-deploy/models/sys/base/resource"
	"go.mongodb.org/mongo-driver/mongo"
)

// Client is the client for worker statuses.
type Client struct {
	Collection *mongo.Collection

	resource.ResourceClient[WorkerStatus]
}

// New creates a new worker status client.
func New() *Client {
	return &Client{
		Collection: db.DB.GetCollection("workerStatus"),

		ResourceClient: resource.ResourceClient[WorkerStatus]{
			Collection:     db.DB.GetCollection("workerStatus"),
			IncludeDeleted: false,
		},
	}
}
