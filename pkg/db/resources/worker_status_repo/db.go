package worker_status_repo

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CreateOrUpdate creates or updates a worker status.
// If the worker status already exists, it will be updated.
func (client *Client) CreateOrUpdate(name, status string) error {
	filter := bson.D{{Key: "name", Value: name}}
	update := bson.D{{Key: "$set", Value: bson.D{
		{Key: "name", Value: name},
		{Key: "status", Value: status},
		{Key: "reportedAt", Value: time.Now()},
	}}}

	_, err := client.Collection.UpdateOne(context.TODO(), filter, update, options.Update().SetUpsert(true))
	if err != nil {
		return err
	}

	return nil
}

// DeleteStale deletes all worker statuses that have not been updated in the last 24 hours.
func (client *Client) DeleteStale() error {
	filter := bson.D{{Key: "reportedAt", Value: bson.D{{Key: "$lt", Value: time.Now().Add(-24 * time.Hour)}}}}
	_, err := client.Collection.DeleteMany(context.TODO(), filter)
	if err != nil {
		return err
	}

	return nil
}
