package resource

import (
	"context"
	"fmt"
	"go-deploy/models"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

// GetByID returns a resource with the given ID.
func (client *ResourceClient[T]) GetByID(id string) (*T, error) {
	return models.GetResource[T](client.Collection, models.GroupFilters(bson.D{{"id", id}}, client.ExtraFilter, client.Search, client.IncludeDeleted), nil)
}

// GetByName returns a resource with the given name.
func (client *ResourceClient[T]) GetByName(name string) (*T, error) {
	return models.GetResource[T](client.Collection, models.GroupFilters(bson.D{{"name", name}}, client.ExtraFilter, client.Search, client.IncludeDeleted), nil)
}

// List returns any resources that match the given filter.
func (client *ResourceClient[T]) List() ([]T, error) {
	return models.ListResources[T](client.Collection, models.GroupFilters(bson.D{}, client.ExtraFilter, client.Search, client.IncludeDeleted), nil, client.Pagination)
}

// ExistsByID returns whether a resource with the given ID exists.
func (client *ResourceClient[T]) ExistsByID(id string) (bool, error) {
	count, err := models.CountResources(client.Collection, models.GroupFilters(bson.D{{"id", id}}, client.ExtraFilter, client.Search, client.IncludeDeleted))
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// ExistsByName returns whether a resource with the given name exists.
func (client *ResourceClient[T]) ExistsByName(name string) (bool, error) {
	count, err := models.CountResources(client.Collection, models.GroupFilters(bson.D{{"name", name}}, client.ExtraFilter, client.Search, client.IncludeDeleted))
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// ExistsWithFilter returns whether a resource with the given filter exists.
func (client *ResourceClient[T]) ExistsWithFilter(filter bson.D) (bool, error) {
	count, err := models.CountResources(client.Collection, models.GroupFilters(filter, client.ExtraFilter, client.Search, client.IncludeDeleted))
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// ExistsAny returns whether any resources exist with the given filter.
func (client *ResourceClient[T]) ExistsAny() (bool, error) {
	count, err := models.CountResources(client.Collection, models.GroupFilters(bson.D{}, client.ExtraFilter, client.Search, client.IncludeDeleted))
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// CreateIfUnique creates a resource with the given ID and resource,
// but only if a resource with the given filter does not already exist.
func (client *ResourceClient[T]) CreateIfUnique(id string, resource *T, filter bson.D) error {
	return models.CreateIfUniqueResource[T](client.Collection, id, resource, models.GroupFilters(filter, client.ExtraFilter, client.Search, client.IncludeDeleted))
}

// UpdateWithBSON updates a resource with the given BSON update.
func (client *ResourceClient[T]) UpdateWithBSON(update bson.D) error {
	return client.UpdateWithBsonByFilter(bson.D{}, update)
}

// UpdateWithBsonByID updates a resource with the given ID and BSON update.
func (client *ResourceClient[T]) UpdateWithBsonByID(id string, update bson.D) error {
	return client.UpdateWithBsonByFilter(bson.D{{"id", id}}, update)
}

// UpdateWithBsonByFilter updates a resource with the given filter and BSON update.
func (client *ResourceClient[T]) UpdateWithBsonByFilter(filter bson.D, update bson.D) error {
	return models.UpdateOneResource(client.Collection, models.GroupFilters(filter, client.ExtraFilter, client.Search, client.IncludeDeleted), update)
}

// UnsetWithBSON unsets the given fields from a resource.
func (client *ResourceClient[T]) UnsetWithBSON(fields ...string) error {
	update := bson.D{
		{"$unset", bson.D{}},
	}

	for _, field := range fields {
		update[0].Value = append(update[0].Value.(bson.D), bson.E{Key: field, Value: ""})
	}

	return client.UpdateWithBSON(update)
}

// UnsetWithBsonByID unsets the given fields from a resource with the given ID.
func (client *ResourceClient[T]) UnsetWithBsonByID(id string, fields ...string) error {
	update := bson.D{
		{"$unset", bson.D{}},
	}

	for _, field := range fields {
		update[0].Value = append(update[0].Value.(bson.D), bson.E{Key: field, Value: ""})
	}

	return client.UpdateWithBsonByID(id, update)
}

// SetWithBSON sets the given fields to the given values in a resource.
func (client *ResourceClient[T]) SetWithBSON(update bson.D) error {
	return client.UpdateWithBSON(bson.D{{"$set", update}})
}

// SetWithBsonByID sets the given fields to the given values in a resource with the given ID.
func (client *ResourceClient[T]) SetWithBsonByID(id string, update bson.D) error {
	return client.UpdateWithBsonByID(id, bson.D{{"$set", update}})
}

// SetWithBsonByFilter sets the given fields to the given values in a resource with the given filter.
func (client *ResourceClient[T]) SetWithBsonByFilter(filter bson.D, update bson.D) error {
	return client.UpdateWithBsonByFilter(filter, bson.D{{"$set", update}})
}

// CountDistinct returns the number of distinct values for the given field.
func (client *ResourceClient[T]) CountDistinct(field string) (int, error) {
	return models.CountDistinctResources(client.Collection, field, models.GroupFilters(bson.D{}, client.ExtraFilter, client.Search, client.IncludeDeleted))
}

// Delete deletes all resources that match the given filter.
// It only sets the deletedAt field to the current time, and
// does not remove the resources from the database.
func (client *ResourceClient[T]) Delete() error {
	_, err := client.Collection.UpdateOne(context.TODO(),
		bson.D{},
		bson.D{
			{"$set", bson.D{{"deletedAt", time.Now()}}},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to delete resources. details: %w", err)
	}

	return nil
}

// DeleteByID deletes a resource with the given ID.
// It only sets the deletedAt field to the current time, and
// does not remove the resource from the database.
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

// Deleted returns whether a resource with the given ID has been deleted.
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

// Get returns a resource with the given filter.
func (client *ResourceClient[T]) Get() (*T, error) {
	return models.GetResource[T](client.Collection, models.GroupFilters(bson.D{}, client.ExtraFilter, client.Search, client.IncludeDeleted), nil)
}

// OnlyID is a type that only contains an ID.
// This is useful when only the ID is needed.
// This should be paired with a projection that only includes the ID,
// such as bson.D{{"id", 1}}.
type OnlyID struct {
	ID string `bson:"id"`
}

// GetID returns the ID of a resource with the given filter.
// It returns a OnlyID type, which only contains the ID.
func (client *ResourceClient[T]) GetID() (*OnlyID, error) {
	projection := bson.D{{"id", 1}}
	return models.GetResource[OnlyID](client.Collection, models.GroupFilters(bson.D{}, client.ExtraFilter, client.Search, client.IncludeDeleted), projection)
}

// ListIDs returns the IDs of all resources that match the given filter.
func (client *ResourceClient[T]) ListIDs() ([]OnlyID, error) {
	projection := bson.D{{"id", 1}}
	return models.ListResources[OnlyID](client.Collection, models.GroupFilters(nil, client.ExtraFilter, client.Search, client.IncludeDeleted), projection, client.Pagination)
}

// GetWithFilterAndProjection returns a resource with the given filter and projection.
// It should be used as a base method for other methods to use, and not called directly.
func (client *ResourceClient[T]) GetWithFilterAndProjection(filter, projection bson.D) (*T, error) {
	return models.GetResource[T](client.Collection, models.GroupFilters(filter, client.ExtraFilter, client.Search, client.IncludeDeleted), projection)
}

// ListWithFilterAndProjection returns any resources that match the given filter and projection.
// It should be used as a base method for other methods to use, and not called directly.
func (client *ResourceClient[T]) ListWithFilterAndProjection(filter, projection bson.D) ([]T, error) {
	return models.ListResources[T](client.Collection, models.GroupFilters(filter, client.ExtraFilter, client.Search, client.IncludeDeleted), projection, client.Pagination)
}

// Count returns the number of resources that match the given filter.
func (client *ResourceClient[T]) Count() (int, error) {
	return models.CountResources(client.Collection, models.GroupFilters(bson.D{}, client.ExtraFilter, client.Search, client.IncludeDeleted))
}
