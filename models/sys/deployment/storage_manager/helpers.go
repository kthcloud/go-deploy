package storage_manager

import (
	"context"
	"errors"
	"fmt"
	"go-deploy/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

func CreateStorageManager(id string, params *CreateParams) (string, error) {
	storageManager := StorageManager{
		ID:        id,
		OwnerID:   params.UserID,
		CreatedAt: time.Now(),
		Zone:      params.Zone,
	}

	_, err := models.StorageManagerCollection.UpdateOne(context.TODO(), bson.D{{"ownerId", params.UserID}}, bson.D{
		{"$setOnInsert", storageManager},
	}, options.Update().SetUpsert(true))
	if err != nil {
		return "", fmt.Errorf("failed to create storage manager. details: %s", err)
	}

	manager, err := getStorageManager(bson.D{{"ownerId", params.UserID}})
	if err != nil {
		return "", fmt.Errorf("failed to fetch storage manager. details: %s", err)
	}

	return manager.ID, nil
}

func DeleteStorageManager(id string) error {
	_, err := models.StorageManagerCollection.DeleteOne(context.TODO(), bson.D{
		{"id", id},
	})
	if err != nil {
		return fmt.Errorf("failed to delete storage manager. details: %s", err)
	}

	return nil
}

func getStorageManager(filter bson.D) (*StorageManager, error) {
	var storageManager StorageManager
	err := models.StorageManagerCollection.FindOne(context.TODO(), filter).Decode(&storageManager)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
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

func GetByOwnerID(ownerID string) (*StorageManager, error) {
	return getStorageManager(bson.D{{"ownerId", ownerID}})
}

func GetAll() ([]StorageManager, error) {
	return GetAllWithFilter(bson.D{})
}

func UpdateByID(id string, update bson.D) error {
	_, err := models.StorageManagerCollection.UpdateOne(context.TODO(), bson.D{{"id", id}}, bson.D{{"$set", update}})
	if err != nil {
		err = fmt.Errorf("failed to update storage manager %s. details: %s", id, err)
		return err
	}
	return nil
}

func UpdateSubsystemByID(id, subsystem string, key string, update interface{}) error {
	subsystemKey := fmt.Sprintf("subsystems.%s.%s", subsystem, key)
	return UpdateByID(id, bson.D{{subsystemKey, update}})
}

func GetAllWithFilter(filter bson.D) ([]StorageManager, error) {
	cursor, err := models.StorageManagerCollection.Find(context.TODO(), filter)

	if err != nil {
		err = fmt.Errorf("failed to fetch all storage manager. details: %s", err)
		log.Println(err)
		return nil, err
	}

	var storageManagers []StorageManager
	for cursor.Next(context.TODO()) {
		var storageManager StorageManager

		err = cursor.Decode(&storageManager)
		if err != nil {
			err = fmt.Errorf("failed to decode storage manager when fetching all storage managers. details: %s", err)
			return nil, err
		}
		storageManagers = append(storageManagers, storageManager)
	}

	return storageManagers, nil
}

func GetWithNoActivities() ([]StorageManager, error) {
	filter := bson.D{
		{
			"activities", bson.M{
			"$size": 0,
		},
		},
	}

	return GetAllWithFilter(filter)
}

func AddActivity(storageManagerID, activity string) error {
	_, err := models.StorageManagerCollection.UpdateOne(context.TODO(),
		bson.D{{"id", storageManagerID}},
		bson.D{{"$addToSet", bson.D{{"activities", activity}}}},
	)
	if err != nil {
		err = fmt.Errorf("failed to add activity %s to storage manager %s. details: %s", activity, storageManagerID, err)
		return err
	}
	return nil
}

func RemoveActivity(storageManagerID, activity string) error {
	_, err := models.StorageManagerCollection.UpdateOne(context.TODO(),
		bson.D{{"id", storageManagerID}},
		bson.D{{"$pull", bson.D{{"activities", activity}}}},
	)
	if err != nil {
		err = fmt.Errorf("failed to remove activity %s from storage manager %s. details: %s", activity, storageManagerID, err)
		return err
	}
	return nil
}
