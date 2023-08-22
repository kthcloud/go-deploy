package storage_manager

import (
	"context"
	"fmt"
	"go-deploy/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func CreateStorageManager(id, ownerID, zone string) (string, error) {
	storageManager := StorageManager{
		ID:        id,
		OwnerID:   ownerID,
		CreatedAt: time.Now(),
		Zone:      zone,
	}

	_, err := models.DeploymentCollection.UpdateOne(context.TODO(), bson.D{{"ownerId", ownerID}}, bson.D{
		{"$setOnInsert", storageManager},
	}, options.Update().SetUpsert(true))
	if err != nil {
		return "", fmt.Errorf("failed to create storage manager. details: %s", err)
	}

	manager, err := getStorageManager(bson.D{{"ownerId", ownerID}})
	if err != nil {
		return "", fmt.Errorf("failed to fetch storage manager. details: %s", err)
	}

	return manager.ID, nil
}

func getStorageManager(filter bson.D) (*StorageManager, error) {
	var storageManager StorageManager
	err := models.DeploymentCollection.FindOne(context.TODO(), filter).Decode(&storageManager)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}

		err = fmt.Errorf("failed to fetch storage manager. details: %s", err)
		invalidStorageManager := StorageManager{}
		return &invalidStorageManager, err
	}

	return &storageManager, err
}

func GetByID(id string) (*StorageManager, error) {
	return getStorageManager(bson.D{{"id", id}})
}

func UpdateByID(id string, update bson.D) error {
	_, err := models.StorageManagerCollection.UpdateOne(context.TODO(), bson.D{{"id", id}}, bson.D{{"$set", update}})
	if err != nil {
		err = fmt.Errorf("failed to update deployment %s. details: %s", id, err)
		return err
	}
	return nil
}

func UpdateSubsystemByID(id, subsystem string, key string, update interface{}) error {
	subsystemKey := fmt.Sprintf("subsystems.%s.%s", subsystem, key)
	return UpdateByID(id, bson.D{{subsystemKey, update}})
}
