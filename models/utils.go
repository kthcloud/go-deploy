package models

import (
	"context"
	"errors"
	"fmt"
	"go-deploy/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"reflect"
	"time"
)

func AddIfNotNil(data bson.M, key string, value interface{}) {
	if value == nil || (reflect.ValueOf(value).Kind() == reflect.Ptr && reflect.ValueOf(value).IsNil()) {
		return
	}
	data[key] = value
}

func addIncludeDeleted(filter bson.D) bson.D {
	if filter == nil {
		filter = bson.D{}
	}

	filter = append(filter, bson.E{Key: "deletedAt", Value: bson.D{{"$ne", time.Time{}}}})

	return filter
}

func addExcludeDeleted(filter bson.D) bson.D {
	if filter == nil {
		filter = bson.D{}
	}

	filter = append(filter, bson.E{Key: "deletedAt", Value: bson.D{{"$in", []interface{}{nil, time.Time{}}}}})

	return filter
}

func GetResource[T any](collection *mongo.Collection, filter bson.D, includeDeleted bool) (*T, error) {
	if includeDeleted {
		filter = addIncludeDeleted(filter)
	} else {
		filter = addExcludeDeleted(filter)
	}

	var resource T
	err := collection.FindOne(context.TODO(), filter).Decode(&resource)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to get resource. details: %w", err)
	}
	return &resource, nil
}

func GetManyResources[T any](collection *mongo.Collection, filter bson.D, includeDeleted bool) ([]T, error) {
	if includeDeleted {
		filter = addIncludeDeleted(filter)
	} else {
		filter = addExcludeDeleted(filter)
	}

	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get resources. details: %w", err)
	}

	defer func(cursor *mongo.Cursor, ctx context.Context) {
		closeErr := cursor.Close(ctx)
		if closeErr != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to close cursor. details: %w", closeErr))
		}
	}(cursor, context.Background())

	var resources []T
	err = cursor.All(context.Background(), &resources)
	if err != nil {
		return nil, fmt.Errorf("failed to get resources. details: %w", err)
	}

	return resources, nil
}

func CountResources(collection *mongo.Collection, filter bson.D, includeDeleted bool) (int, error) {
	if includeDeleted {
		filter = addIncludeDeleted(filter)
	} else {
		filter = addExcludeDeleted(filter)
	}

	count, err := collection.CountDocuments(context.Background(), filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count resources. details: %w", err)
	}
	return int(count), nil
}
