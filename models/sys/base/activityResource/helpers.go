package activityResource

import (
	"context"
	"fmt"
	"go-deploy/models"
	activityModels "go-deploy/models/sys/activity"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

// AddActivity adds an activity to the resource with the given ID.
// If the activity already exists, it will be overwritten.
func (client *ActivityResourceClient[T]) AddActivity(id, activity string) error {
	_, err := client.Collection.UpdateOne(context.TODO(),
		bson.D{{"id", id}},
		bson.D{{"$set", bson.D{
			{"activities." + activity, activityModels.Activity{
				Name:      activity,
				CreatedAt: time.Now(),
			}},
		}}},
	)
	if err != nil {
		err = fmt.Errorf("failed to add activity %s to resource %s. details: %w", activity, id, err)
		return err
	}
	return nil
}

// RemoveActivity removes an activity from the resource with the given ID.
// If the activity does not exist, nothing will happen.
func (client *ActivityResourceClient[T]) RemoveActivity(id, activity string) error {
	_, err := client.Collection.UpdateOne(context.TODO(),
		bson.D{{"id", id}},
		bson.D{{"$unset", bson.D{
			{"activities." + activity, ""},
		}}},
	)
	if err != nil {
		err = fmt.Errorf("failed to remove activity %s from resource %s. details: %w", activity, id, err)
		return err
	}
	return nil
}

// ClearActivities clears all activities from the resource with the given ID.
func (client *ActivityResourceClient[T]) ClearActivities(id string) error {
	_, err := client.Collection.UpdateOne(context.TODO(),
		bson.D{{"id", id}},
		bson.D{{"$set", bson.D{
			{"activities", make(map[string]activityModels.Activity)},
		}}},
	)
	if err != nil {
		err = fmt.Errorf("failed to clear activities from resource %s. details: %w", id, err)
		return err
	}

	return nil
}

// IsDoingActivity returns whether the resource with the given ID is doing the given activity.
func (client *ActivityResourceClient[T]) IsDoingActivity(id, activity string) (bool, error) {
	filter := bson.D{
		{"id", id},
		{"activities." + activity, bson.M{
			"$exists": true,
		}},
	}

	count, err := models.CountResources(client.Collection, models.GroupFilters(filter, client.ExtraFilter, client.Search, false))
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
