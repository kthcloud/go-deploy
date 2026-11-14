package db

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/kthcloud/go-deploy/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Resource any

type onlyID struct {
	ID string `bson:"id"`
}

var ErrUniqueConstraint = errors.New("unique constraint error")

func AddIfNotNil(data *bson.D, key string, value any) {
	if value == nil || (reflect.ValueOf(value).Kind() == reflect.Pointer && reflect.ValueOf(value).IsNil()) {
		return
	}
	*data = append(*data, bson.E{Key: key, Value: value})
}

func Add(data *bson.D, key string, value any) {
	*data = append(*data, bson.E{Key: key, Value: value})
}

func GroupFilters(filter bson.D, extraFilter bson.M, searchParams *SearchParams, includeDeleted bool) bson.D {
	// deleted filter
	if !includeDeleted {
		filter = AddExcludeDeletedFilter(filter)
	}

	// extra filter
	if extraFilter != nil {
		filter = bson.D{{Key: "$and", Value: bson.A{filter, extraFilter}}}
	}

	// search filter
	if searchParams != nil {
		searchFilter := bson.A{}
		for _, field := range searchParams.Fields {
			pattern := fmt.Sprintf(".*%s.*", searchParams.Query)
			searchFilter = append(searchFilter, bson.M{field: bson.M{"$regex": pattern, "$options": "i"}})
		}

		return bson.D{{Key: "$and", Value: bson.A{filter, bson.D{{Key: "$or", Value: searchFilter}}}}}
	} else {
		return filter
	}
}

func AddExcludeDeletedFilter(filter bson.D) bson.D {
	newFilter := filter

	newFilter = append(newFilter, bson.E{Key: "deletedAt", Value: bson.D{{Key: "$in", Value: bson.A{nil, time.Time{}}}}})

	return newFilter
}

func normalizeCtx(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

func GetResource[T Resource](ctx context.Context, collection *mongo.Collection, filter bson.D, projection bson.D) (*T, error) {
	ctx = normalizeCtx(ctx)
	findOptions := options.FindOne()
	if projection != nil {
		findOptions.SetProjection(projection)
	}

	var res T
	err := collection.FindOne(ctx, filter, findOptions).Decode(&res)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to get model. details: %w", err)
	}
	return &res, nil
}

func ListResources[T any](ctx context.Context, collection *mongo.Collection, filter bson.D, projection bson.D, pagination *Pagination, sortBy *SortBy) ([]T, error) {
	ctx = normalizeCtx(ctx)
	findOptions := &options.FindOptions{}
	if pagination != nil {
		limit := int64(pagination.PageSize)

		var skip int64
		if pagination.Page > 0 {
			skip = int64(pagination.Page * pagination.PageSize)
		}

		findOptions.SetLimit(limit)
		findOptions.SetSkip(skip)
	}

	if sortBy != nil {
		findOptions.SetSort(bson.D{{Key: sortBy.Field, Value: sortBy.Order}})
	}

	if projection != nil {
		findOptions.SetProjection(projection)
	}

	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to get resources. details: %w", err)
	}

	defer func(cursor *mongo.Cursor, ctx context.Context) {
		closeErr := cursor.Close(ctx)
		if closeErr != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to close cursor. details: %w", closeErr))
		}
	}(cursor, ctx)

	var res []T
	err = cursor.All(ctx, &res)
	if err != nil {
		return nil, fmt.Errorf("failed to get resources. details: %w", err)
	}

	return res, nil
}

func CountResources(ctx context.Context, collection *mongo.Collection, filter bson.D) (int, error) {
	count, err := collection.CountDocuments(normalizeCtx(ctx), filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count resources. details: %w", err)
	}
	return int(count), nil
}

func CountDistinctResources(ctx context.Context, collection *mongo.Collection, field string, filter bson.D) (int, error) {
	count, err := collection.Distinct(normalizeCtx(ctx), field, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count resources. details: %w", err)
	}
	return len(count), nil
}

func CreateIfUniqueResource[T Resource](ctx context.Context, collection *mongo.Collection, id string, data *T, filter bson.D) error {
	ctx = normalizeCtx(ctx)
	result, err := collection.UpdateOne(ctx, filter, bson.D{
		{Key: "$setOnInsert", Value: data},
	}, options.Update().SetUpsert(true))
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return ErrUniqueConstraint
		}

		return fmt.Errorf("failed to create unique model. details: %w", err)
	}

	if result.UpsertedCount == 0 {
		if result.MatchedCount == 1 {
			fetched, err := GetResource[onlyID](ctx, collection, filter, nil)
			if err != nil {
				return err
			}

			if fetched == nil {
				utils.PrettyPrintError(fmt.Errorf("failed to fetch model after creation. Assuming it was deleted"))
				return nil
			}

			if fetched.ID == id {
				return nil
			}
		}

		return ErrUniqueConstraint
	}

	return nil
}

func UpdateOneResource(ctx context.Context, collection *mongo.Collection, filter bson.D, update bson.D) error {
	_, err := collection.UpdateOne(normalizeCtx(ctx), filter, update)
	if err != nil {
		return fmt.Errorf("failed to update model. details: %w", err)
	}

	return nil
}

func DeleteResources(ctx context.Context, collection *mongo.Collection, filter bson.D) error {
	_, err := collection.DeleteMany(normalizeCtx(ctx), filter)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil
		}

		return fmt.Errorf("failed to delete resources. details: %w", err)
	}

	return nil
}
