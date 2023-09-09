package activityResource

import (
	"context"
	"fmt"
	"go-deploy/models"
	"go.mongodb.org/mongo-driver/bson"
)

func (client ActivityResourceClient[T]) GetByActivity(activity string) ([]T, error) {
	filter := bson.D{{
		"activities", bson.M{
			"$in": bson.A{activity},
		},
	}}

	return models.GetManyResources[T](client.Collection, filter, false)
}

func (client ActivityResourceClient[T]) GetWithNoActivities() ([]T, error) {
	filter := bson.D{{
		"activities", bson.M{
			"$size": 0,
		},
	}}

	return models.GetManyResources[T](client.Collection, filter, false)
}

func (client ActivityResourceClient[T]) AddActivity(id, activity string) error {
	_, err := client.Collection.UpdateOne(context.TODO(),
		bson.D{{"id", id}},
		bson.D{{"$addToSet", bson.D{{"activities", activity}}}},
	)
	if err != nil {
		err = fmt.Errorf("failed to add activity %s to resource %s. details: %w", activity, id, err)
		return err
	}
	return nil
}

func (client ActivityResourceClient[T]) RemoveActivity(id, activity string) error {
	_, err := client.Collection.UpdateOne(context.TODO(),
		bson.D{{"id", id}},
		bson.D{{"$pull", bson.D{{"activities", activity}}}},
	)
	if err != nil {
		err = fmt.Errorf("failed to remove activity %s from resource %s. details: %w", activity, id, err)
		return err
	}
	return nil
}

func (client ActivityResourceClient[T]) ClearActivities(id string) error {
	_, err := client.Collection.UpdateOne(context.TODO(),
		bson.D{{"id", id}},
		bson.D{{"$set", bson.D{{"activities", bson.A{}}}}},
	)
	if err != nil {
		err = fmt.Errorf("failed to clear activities from resource %s. details: %w", id, err)
		return err
	}

	return nil
}

func (client ActivityResourceClient[T]) DoingActivity(id, activity string) (bool, error) {
	filter := bson.D{
		{"id", id},
		{"activities", activity},
	}

	count, err := models.CountResources(client.Collection, filter, false)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
