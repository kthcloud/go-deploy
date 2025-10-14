package worker_status_repo

import (
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db"
	"github.com/kthcloud/go-deploy/pkg/db/resources/base_clients"
	"go.mongodb.org/mongo-driver/mongo"
)

// Client is the client for worker statuses.
type Client struct {
	Collection *mongo.Collection

	base_clients.ResourceClient[model.WorkerStatus]
}

// New creates a new worker status client.
func New() *Client {
	return &Client{
		Collection: db.DB.GetCollection("workerStatus"),

		ResourceClient: base_clients.ResourceClient[model.WorkerStatus]{
			Collection:     db.DB.GetCollection("workerStatus"),
			IncludeDeleted: false,
		},
	}
}
