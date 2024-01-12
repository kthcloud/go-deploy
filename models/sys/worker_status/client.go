package worker_status

import (
	"go-deploy/models/db"
	"go-deploy/models/sys/base/resource"
	"go.mongodb.org/mongo-driver/mongo"
)

type Client struct {
	Collection *mongo.Collection

	resource.ResourceClient[WorkerStatus]
}

func New() *Client {
	return &Client{
		Collection: db.DB.GetCollection("workerStatus"),

		ResourceClient: resource.ResourceClient[WorkerStatus]{
			Collection:     db.DB.GetCollection("workerStatus"),
			IncludeDeleted: false,
		},
	}
}
