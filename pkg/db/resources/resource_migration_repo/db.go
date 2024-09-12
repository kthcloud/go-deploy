package resource_migration_repo

import (
	"fmt"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db"
	"go.mongodb.org/mongo-driver/bson"
	"time"
)

func (c *Client) Create(id, userID, resourceID, migrationType, resourceType string, code *string, status string, params interface{}) (*model.ResourceMigration, error) {
	migration := &model.ResourceMigration{
		ID:         id,
		ResourceID: resourceID,
		UserID:     userID,

		Type:         migrationType,
		ResourceType: resourceType,
		Code:         code,

		Status:    status,
		CreatedAt: time.Now(),
	}

	switch migrationType {
	case model.ResourceMigrationTypeUpdateOwner:
		updateOwnerParams, ok := params.(*model.ResourceMigrationUpdateOwnerParams)
		if !ok {
			return nil, fmt.Errorf("bad params for migration type %s", migrationType)
		}

		migration.UpdateOwner = &model.ResourceMigrationUpdateOwner{
			NewOwnerID: updateOwnerParams.NewOwnerID,
			OldOwnerID: updateOwnerParams.OldOwnerID,
		}

	default:
		return nil, fmt.Errorf("bad migration type %s", migrationType)
	}

	err := c.CreateIfUnique(id, migration, bson.D{{"resourceId", resourceID}})
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
	return c.SetWithBsonByID(id, update)
}
