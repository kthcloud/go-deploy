package base_clients

import (
	"context"
	"fmt"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

// ActivityResourceClient is a type of base client that adds methods to manage activities in the database.
type ActivityResourceClient[T any] struct {
	Collection  *mongo.Collection
	ExtraFilter bson.M
	Pagination  *db.Pagination
	Search      *db.SearchParams
}

// AddExtraFilter adds an extra filter to the client.
func (client *ActivityResourceClient[T]) AddExtraFilter(filter bson.D) *ActivityResourceClient[T] {
	if client.ExtraFilter == nil {
		client.ExtraFilter = bson.M{
			"$and": bson.A{},
		}
	}

	client.ExtraFilter["$and"] = append(client.ExtraFilter["$and"].(bson.A), filter)

	return client
}

// AddActivity adds an activity to the model with the given ID.
// If the activity already exists, it will be overwritten.
func (client *ActivityResourceClient[T]) AddActivity(id, activity string) error {
	_, err := client.Collection.UpdateOne(context.TODO(),
		bson.D{{"id", id}},
		bson.D{{"$set", bson.D{
			{"activities." + activity, model.Activity{
				Name:      activity,
				CreatedAt: time.Now(),
			}},
		}}},
	)
	if err != nil {
		err = fmt.Errorf("failed to add activity %s to model %s. details: %w", activity, id, err)
		return err
	}
	return nil
}

// RemoveActivity removes an activity from the model with the given ID.
// If the activity does not exist, nothing will happen.
func (client *ActivityResourceClient[T]) RemoveActivity(id, activity string) error {
	_, err := client.Collection.UpdateOne(context.TODO(),
		bson.D{{"id", id}},
		bson.D{{"$unset", bson.D{
			{"activities." + activity, ""},
		}}},
	)
	if err != nil {
		err = fmt.Errorf("failed to remove activity %s from model %s. details: %w", activity, id, err)
		return err
	}
	return nil
}

// ClearActivities clears all activities from the model with the given ID.
func (client *ActivityResourceClient[T]) ClearActivities(id string) error {
	_, err := client.Collection.UpdateOne(context.TODO(),
		bson.D{{"id", id}},
		bson.D{{"$set", bson.D{
			{"activities", make(map[string]model.Activity)},
		}}},
	)
	if err != nil {
		err = fmt.Errorf("failed to clear activities from model %s. details: %w", id, err)
		return err
	}

	return nil
}

// IsDoingActivity returns whether the model with the given ID is doing the given activity.
func (client *ActivityResourceClient[T]) IsDoingActivity(id, activity string) (bool, error) {
	filter := bson.D{
		{"id", id},
		{"activities." + activity, bson.M{
			"$exists": true,
		}},
	}

	count, err := db.CountResources(client.Collection, db.GroupFilters(filter, client.ExtraFilter, client.Search, false))
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
