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

type Resource interface {
}

type onlyID struct {
	ID string `bson:"id"`
}

type SearchParams struct {
	Query  string   `json:"query"`
	Fields []string `json:"fields"`
}

var UniqueConstraintErr = errors.New("unique constraint error")

func AddIfNotNil(data *bson.D, key string, value interface{}) {
	if value == nil || (reflect.ValueOf(value).Kind() == reflect.Ptr && reflect.ValueOf(value).IsNil()) {
		return
	}
	*data = append(*data, bson.E{Key: key, Value: value})
}

func addExcludeDeleted(filter bson.D) bson.D {
	newFilter := filter

	newFilter = append(newFilter, bson.E{Key: "deletedAt", Value: bson.D{{"$in", []interface{}{nil, time.Time{}}}}})

	return newFilter
}

func GetResource[T Resource](collection *mongo.Collection, filter bson.D, includeDeleted bool, extraFilter bson.D) (*T, error) {
	var res T
	err := collection.FindOne(context.TODO(), groupFilters(filter, extraFilter, nil, includeDeleted)).Decode(&res)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}

		return nil, fmt.Errorf("failed to get resource. details: %w", err)
	}
	return &res, nil
}

func ListResources[T any](collection *mongo.Collection, filter bson.D, includeDeleted bool, pagination *base.Pagination, extraFilter bson.D, search *SearchParams) ([]T, error) {
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

	cursor, err := collection.Find(context.Background(), groupFilters(filter, extraFilter, search, includeDeleted), findOptions)
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

func CountResources(collection *mongo.Collection, filter bson.D, includeDeleted bool, extraFilter bson.D) (int, error) {
	count, err := collection.CountDocuments(context.Background(), groupFilters(filter, extraFilter, nil, includeDeleted))
	if err != nil {
		return 0, fmt.Errorf("failed to count resources. details: %w", err)
	}
	return int(count), nil
}

func CountDistinctResources(collection *mongo.Collection, field string, filter bson.D, includeDeleted bool, extraFilter bson.D) (int, error) {
	count, err := collection.Distinct(context.Background(), field, groupFilters(filter, extraFilter, nil, includeDeleted))
	if err != nil {
		return 0, fmt.Errorf("failed to count resources. details: %w", err)
	}
	return len(count), nil
}

func CreateIfUniqueResource[T Resource](collection *mongo.Collection, id string, data *T, filter bson.D, includeDeleted bool, extraFilter bson.D) error {
	result, err := collection.UpdateOne(context.TODO(), groupFilters(filter, extraFilter, nil, includeDeleted), bson.D{
		{"$setOnInsert", data},
	}, options.Update().SetUpsert(true))
	if err != nil {
		return fmt.Errorf("failed to create unique resource. details: %w", err)
	}

	if result.UpsertedCount == 0 {
		if result.MatchedCount == 1 {
			fetched, err := GetResource[onlyID](collection, filter, includeDeleted, extraFilter)
			if err != nil {
				return err
			}

			if fetched == nil {
				utils.PrettyPrintError(fmt.Errorf("failed to fetch resource after creation. assuming it was deleted"))
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

func UpdateOneResource(collection *mongo.Collection, filter bson.D, update bson.D, includeDeleted bool, extraFilter bson.D) error {
	_, err := collection.UpdateOne(context.Background(), groupFilters(filter, extraFilter, nil, includeDeleted), update)
	if err != nil {
		return fmt.Errorf("failed to update resource. details: %w", err)
	}

	return nil
}

func groupFilters(filter bson.D, extraFilter bson.D, searchParams *SearchParams, includeDeleted bool) bson.M {
	// deleted filter
	if !includeDeleted {
		filter = addExcludeDeleted(filter)
	}

	// extra filter
	filter = append(filter, extraFilter...)

	// search filter
	if searchParams != nil {
		searchFilter := bson.A{}
		for _, field := range searchParams.Fields {
			pattern := fmt.Sprintf(".*%s.*", searchParams.Query)
			searchFilter = append(searchFilter, bson.M{field: bson.M{"$regex": pattern, "$options": "i"}})
		}

		return bson.M{
			"$and": bson.A{
				filter,
				bson.M{
					"$or": searchFilter,
				},
			},
		}

	} else {
		return bson.M{"$and": bson.A{filter}}
	}
}
