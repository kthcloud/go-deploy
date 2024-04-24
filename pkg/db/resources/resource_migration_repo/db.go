package resource_migration_repo

import (
	"context"
	"fmt"
	"go-deploy/models/model"
	"go-deploy/pkg/db"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

func (c *Client) Create(id, userID, resourceID, migrationType, resourceType, status string, params interface{}) (*model.ResourceMigration, error) {
	migration := &model.ResourceMigration{
		ID:         id,
		ResourceID: resourceID,
		UserID:     userID,

		Type:         migrationType,
		ResourceType: resourceType,

		Status:    status,
		CreatedAt: time.Now(),

		Params: params,
	}

	_, err := c.Collection.InsertOne(context.TODO(), migration)
	if err != nil {
		return nil, fmt.Errorf("failed to create migration. details: %w", err)
	}

	return migration, nil
}

// UpdateWithParams updates the migration with the given id with the given params.
// These are usually parsed from a DTO.
func (c *Client) UpdateWithParams(id string, params *model.ResourceMigrationUpdateParams) error {
	update := bson.D{}
	db.AddIfNotNil(&update, "status", params.Status)
	return c.UpdateWithBsonByID(id, update)
}
