package resource

import (
	"context"
	"fmt"
	"go-deploy/models"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

func (client *ResourceClient[T]) GetByID(id string) (*T, error) {
	return models.GetResource[T](client.Collection, models.GroupFilters(bson.D{{"id", id}}, client.ExtraFilter, client.Search, client.IncludeDeleted))
}

func (client *ResourceClient[T]) GetByName(name string) (*T, error) {
	return models.GetResource[T](client.Collection, models.GroupFilters(bson.D{{"name", name}}, client.ExtraFilter, client.Search, client.IncludeDeleted))
}

func (client *ResourceClient[T]) List() ([]T, error) {
	return models.ListResources[T](client.Collection, models.GroupFilters(bson.D{}, client.ExtraFilter, client.Search, client.IncludeDeleted), client.Pagination)
}

func (client *ResourceClient[T]) ExistsByID(id string) (bool, error) {
	count, err := models.CountResources(client.Collection, models.GroupFilters(bson.D{{"id", id}}, client.ExtraFilter, client.Search, client.IncludeDeleted))
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (client *ResourceClient[T]) ExistsByName(name string) (bool, error) {
	count, err := models.CountResources(client.Collection, models.GroupFilters(bson.D{{"name", name}}, client.ExtraFilter, client.Search, client.IncludeDeleted))
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (client *ResourceClient[T]) ExistsWithFilter(filter bson.D) (bool, error) {
	count, err := models.CountResources(client.Collection, models.GroupFilters(filter, client.ExtraFilter, client.Search, client.IncludeDeleted))
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (client *ResourceClient[T]) ExistsAny() (bool, error) {
	count, err := models.CountResources(client.Collection, models.GroupFilters(bson.D{}, client.ExtraFilter, client.Search, client.IncludeDeleted))
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (client *ResourceClient[T]) CreateIfUnique(id string, resource *T, filter bson.D) error {
	return models.CreateIfUniqueResource[T](client.Collection, id, resource, models.GroupFilters(filter, client.ExtraFilter, client.Search, client.IncludeDeleted))
}

func (client *ResourceClient[T]) UpdateWithBson(update bson.D) error {
	return client.UpdateWithBsonByFilter(bson.D{}, update)
}

func (client *ResourceClient[T]) UpdateWithBsonByID(id string, update bson.D) error {
	return client.UpdateWithBsonByFilter(bson.D{{"id", id}}, update)
}

func (client *ResourceClient[T]) UpdateWithBsonByFilter(filter bson.D, update bson.D) error {
	return models.UpdateOneResource(client.Collection, models.GroupFilters(filter, client.ExtraFilter, client.Search, client.IncludeDeleted), update)
}

func (client *ResourceClient[T]) SetWithBson(update bson.D) error {
	return client.UpdateWithBson(bson.D{{"$set", update}})
}

func (client *ResourceClient[T]) SetWithBsonByID(id string, update bson.D) error {
	return client.UpdateWithBsonByID(id, bson.D{{"$set", update}})
}

func (client *ResourceClient[T]) SetWithBsonByFilter(filter bson.D, update bson.D) error {
	return client.UpdateWithBsonByFilter(filter, bson.D{{"$set", update}})
}

func (client *ResourceClient[T]) CountDistinct(field string) (int, error) {
	return models.CountDistinctResources(client.Collection, field, models.GroupFilters(bson.D{}, client.ExtraFilter, client.Search, client.IncludeDeleted))
}

func (client *ResourceClient[T]) DeleteByID(id string) error {
	_, err := client.Collection.UpdateOne(context.TODO(),
		bson.D{{"id", id}},
		bson.D{
			{"$set", bson.D{{"deletedAt", time.Now()}}},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to delete resource %s. details: %w", id, err)
	}

	return nil
}

func (client *ResourceClient[T]) Deleted(id string) (bool, error) {
	filter := bson.D{
		{"id", id},
		{"deletedAt", bson.M{"$nin": []interface{}{nil, time.Time{}}}},
	}
	count, err := models.CountResources(client.Collection, models.GroupFilters(filter, client.ExtraFilter, client.Search, true))
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (client *ResourceClient[T]) Get() (*T, error) {
	return models.GetResource[T](client.Collection, models.GroupFilters(bson.D{}, client.ExtraFilter, client.Search, client.IncludeDeleted))
}

type OnlyID struct {
	ID string `bson:"id"`
}

func (client *ResourceClient[T]) GetID() (*OnlyID, error) {
	return models.GetResource[OnlyID](client.Collection, models.GroupFilters(bson.D{}, client.ExtraFilter, client.Search, client.IncludeDeleted))
}

func (client *ResourceClient[T]) GetWithFilter(filter bson.D) (*T, error) {
	return models.GetResource[T](client.Collection, models.GroupFilters(filter, client.ExtraFilter, client.Search, client.IncludeDeleted))
}

func (client *ResourceClient[T]) ListWithFilter(filter bson.D) ([]T, error) {
	return models.ListResources[T](client.Collection, models.GroupFilters(filter, client.ExtraFilter, client.Search, client.IncludeDeleted), client.Pagination)
}

func (client *ResourceClient[T]) Count() (int, error) {
	return models.CountResources(client.Collection, models.GroupFilters(bson.D{}, client.ExtraFilter, client.Search, client.IncludeDeleted))
}
