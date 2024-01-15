package worker_status

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

// CreateOrUpdate creates or updates a worker status.
// If the worker status already exists, it will be updated.
func (client *Client) CreateOrUpdate(name, status string) error {
	filter := bson.D{{"name", name}}
	update := bson.D{{"$set", bson.D{
		{"name", name},
		{"status", status},
		{"reportedAt", time.Now()},
	}}}

	_, err := client.Collection.UpdateOne(context.TODO(), filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}

	return nil
}
