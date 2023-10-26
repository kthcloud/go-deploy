package storage_manager

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func (client *Client) CreateStorageManager(id string, params *CreateParams) (string, error) {
	storageManager := StorageManager{
		ID:        id,
		OwnerID:   params.UserID,
		CreatedAt: time.Now(),
		Zone:      params.Zone,
	}

	_, err := client.Collection.UpdateOne(context.TODO(), bson.D{{"ownerId", params.UserID}}, bson.D{
		{"$setOnInsert", storageManager},
	}, options.Update().SetUpsert(true))
	if err != nil {
		return "", fmt.Errorf("failed to create storage manager. details: %w", err)
	}

	manager, err := client.GetWithFilter(bson.D{{"ownerId", params.UserID}})
	if err != nil {
		return "", fmt.Errorf("failed to fetch storage manager. details: %w", err)
	}

	return manager.ID, nil
}

func (client *Client) UpdateSubsystemByID(id, key string, update interface{}) error {
	subsystemKey := fmt.Sprintf("subsystems.%s", key)
	return client.SetWithBsonByID(id, bson.D{{subsystemKey, update}})
}
