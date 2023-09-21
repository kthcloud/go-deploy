package models

import (
	"context"
	"errors"
	"fmt"
	"go-deploy/models/sys/base"
	"go-deploy/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"reflect"
	"time"
)

func AddIfNotNil(data bson.M, key string, value interface{}) {
	if value == nil || (reflect.ValueOf(value).Kind() == reflect.Ptr && reflect.ValueOf(value).IsNil()) {
		return
	}
	data[key] = value
}

func addExcludeDeleted(filter bson.D) bson.D {
	if filter == nil {
		filter = bson.D{}
	}

	filter = append(filter, bson.E{Key: "deletedAt", Value: bson.D{{"$in", []interface{}{nil, time.Time{}}}}})

	return filter
}

func GetResource[T any](collection *mongo.Collection, filter bson.D, includeDeleted bool, extraFilter *bson.D) (*T, error) {
	if !includeDeleted {
		filter = addExcludeDeleted(filter)
	}

	if extraFilter != nil {
		filter = append(filter, *extraFilter...)
	}

	var res T
	err := collection.FindOne(context.TODO(), filter).Decode(&res)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to get resource. details: %w", err)
	}
	return &res, nil
}

func GetManyResources[T any](collection *mongo.Collection, filter bson.D, includeDeleted bool, pagination *base.Pagination, extraFilter *bson.D) ([]T, error) {
	if !includeDeleted {
		filter = addExcludeDeleted(filter)
	}

	if extraFilter != nil {
		filter = append(filter, *extraFilter...)
	}

	var findOptions *options.FindOptions
	if pagination != nil {
		limit := int64(pagination.PageSize)

		var skip int64
		if pagination.Page > 0 {
			skip = int64(pagination.Page * pagination.PageSize)
		}

		findOptions = &options.FindOptions{
			Limit: &limit,
			Skip:  &skip,
		}
	}

	cursor, err := collection.Find(context.Background(), filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to get resources. details: %w", err)
	}

	defer func(cursor *mongo.Cursor, ctx context.Context) {
		closeErr := cursor.Close(ctx)
		if closeErr != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to close cursor. details: %w", closeErr))
		}
	}(cursor, context.Background())

	var res []T
	err = cursor.All(context.Background(), &res)
	if err != nil {
		return nil, fmt.Errorf("failed to get resources. details: %w", err)
	}

	return res, nil
}

func CountResources(collection *mongo.Collection, filter bson.D, includeDeleted bool, extraFilter *bson.D) (int, error) {
	if !includeDeleted {
		filter = addExcludeDeleted(filter)
	}

	if extraFilter != nil {
		filter = append(filter, *extraFilter...)
	}

	count, err := collection.CountDocuments(context.Background(), filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count resources. details: %w", err)
	}
	return int(count), nil
}

func UpdateOneResource(collection *mongo.Collection, filter bson.D, update bson.D, includeDeleted bool, extraFilter *bson.D) error {
	if !includeDeleted {
		filter = addExcludeDeleted(filter)
	}

	if extraFilter != nil {
		filter = append(filter, *extraFilter...)
	}

	_, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return fmt.Errorf("failed to update resource. details: %w", err)
	}

	return nil
}
