package db

import (
	"context"
	"errors"
	"fmt"
	"github.com/kthcloud/go-deploy/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"reflect"
	"time"
)

type Resource interface {
}

type onlyID struct {
	ID string `bson:"id"`
}

var UniqueConstraintErr = errors.New("unique constraint error")

func AddIfNotNil(data *bson.D, key string, value interface{}) {
	if value == nil || (reflect.ValueOf(value).Kind() == reflect.Ptr && reflect.ValueOf(value).IsNil()) {
		return
	}
	*data = append(*data, bson.E{Key: key, Value: value})
}

func Add(data *bson.D, key string, value interface{}) {
	*data = append(*data, bson.E{Key: key, Value: value})
}

func GroupFilters(filter bson.D, extraFilter bson.M, searchParams *SearchParams, includeDeleted bool) bson.D {
	// deleted filter
	if !includeDeleted {
		filter = AddExcludeDeletedFilter(filter)
	}

	// extra filter
	if extraFilter != nil {
		filter = bson.D{{"$and", bson.A{filter, extraFilter}}}
	}

	// search filter
	if searchParams != nil {
		searchFilter := bson.A{}
		for _, field := range searchParams.Fields {
			pattern := fmt.Sprintf(".*%s.*", searchParams.Query)
			searchFilter = append(searchFilter, bson.M{field: bson.M{"$regex": pattern, "$options": "i"}})
		}

		return bson.D{{"$and", bson.A{filter, bson.D{{"$or", searchFilter}}}}}
	} else {
		return filter
	}
}

func AddExcludeDeletedFilter(filter bson.D) bson.D {
	newFilter := filter

	newFilter = append(newFilter, bson.E{Key: "deletedAt", Value: bson.D{{"$in", bson.A{nil, time.Time{}}}}})

	return newFilter
}

func GetResource[T Resource](collection *mongo.Collection, filter bson.D, projection bson.D) (*T, error) {
	findOptions := options.FindOne()
	if projection != nil {
		findOptions.SetProjection(projection)
	}

	var res T
	err := collection.FindOne(context.Background(), filter, findOptions).Decode(&res)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to get model. details: %w", err)
	}
	return &res, nil
}

func ListResources[T any](collection *mongo.Collection, filter bson.D, projection bson.D, pagination *Pagination, sortBy *SortBy) ([]T, error) {
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
		findOptions.SetSort(bson.D{{sortBy.Field, sortBy.Order}})
	}

	if projection != nil {
		findOptions.SetProjection(projection)
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

func CountResources(collection *mongo.Collection, filter bson.D) (int, error) {
	count, err := collection.CountDocuments(context.Background(), filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count resources. details: %w", err)
	}
	return int(count), nil
}

func CountDistinctResources(collection *mongo.Collection, field string, filter bson.D) (int, error) {
	count, err := collection.Distinct(context.Background(), field, filter)
	if err != nil {
		return 0, fmt.Errorf("failed to count resources. details: %w", err)
	}
	return len(count), nil
}

func CreateIfUniqueResource[T Resource](collection *mongo.Collection, id string, data *T, filter bson.D) error {
	result, err := collection.UpdateOne(context.TODO(), filter, bson.D{
		{"$setOnInsert", data},
	}, options.Update().SetUpsert(true))
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return UniqueConstraintErr
		}

		return fmt.Errorf("failed to create unique model. details: %w", err)
	}

	if result.UpsertedCount == 0 {
		if result.MatchedCount == 1 {
			fetched, err := GetResource[onlyID](collection, filter, nil)
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

		return UniqueConstraintErr
	}

	return nil
}

func UpdateOneResource(collection *mongo.Collection, filter bson.D, update bson.D) error {
	_, err := collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return fmt.Errorf("failed to update model. details: %w", err)
	}

	return nil
}

func DeleteResources(collection *mongo.Collection, filter bson.D) error {
	_, err := collection.DeleteMany(context.Background(), filter)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil
		}

		return fmt.Errorf("failed to delete resources. details: %w", err)
	}

	return nil
}
