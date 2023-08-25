package deployment_service

import (
	"fmt"
	"go-deploy/models/sys/deployment/storage_manager"
	"go-deploy/service"
	"go-deploy/service/deployment_service/k8s_service"
	"log"
)

func GetAllStorageManagers(auth *service.AuthInfo) ([]storage_manager.StorageManager, error) {
	if auth.IsAdmin {
		return storage_manager.GetAll()
	}

	ownerStorageManager, err := storage_manager.GetByOwnerID(auth.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch storage manager. details: %s", err)
	}

	if ownerStorageManager == nil {
		return nil, nil
	}

	return []storage_manager.StorageManager{*ownerStorageManager}, nil
}

func GetStorageManagerByOwnerID(ownerID string, auth *service.AuthInfo) (*storage_manager.StorageManager, error) {
	if ownerID != auth.UserID && !auth.IsAdmin {
		return nil, nil
	}

	storageManager, err := storage_manager.GetByOwnerID(ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch storage manager. details: %s", err)
	}

	if storageManager == nil {
		return nil, nil
	}

	return storageManager, nil
}

func GetStorageManagerByID(id string, auth *service.AuthInfo) (*storage_manager.StorageManager, error) {
	storageManager, err := storage_manager.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch storage manager. details: %s", err)
	}

	if storageManager == nil || (storageManager.OwnerID != auth.UserID && !auth.IsAdmin) {
		return nil, nil
	}

	return storageManager, nil
}

func CreateStorageManager(id string, params *storage_manager.CreateParams) error {
	makeErr := func(err error) error {
		return fmt.Errorf("failed to create storage manager. details: %s", err)
	}

	if params == nil {
		return makeErr(fmt.Errorf("params is nil"))
	}

	id, err := storage_manager.CreateStorageManager(id, params)
	if err != nil {
		return makeErr(err)
	}

	err = k8s_service.CreateStorageManager(id, params)
	if err != nil {
		return makeErr(err)
	}

	return nil
}

func DeleteStorageManager(id string) error {
	makeErr := func(err error) error {
		return fmt.Errorf("failed to delete storage manager. details: %s", err)
	}

	log.Println("deleting storage manager", id)

	err := k8s_service.DeleteStorageManager(id)
	if err != nil {
		return makeErr(err)
	}

	err = storage_manager.DeleteStorageManager(id)
	if err != nil {
		return makeErr(err)
	}

	// TODO:
	// New idea! check if everything is deleted at the end, if not, fail the function
	// this will solve the repair <-> delete race, as it will eventually converge as long as the repair
	// respects the beingDeleted activity

	return nil
}

func RepairStorageManager(id string) error {
	makeErr := func(err error) error {
		return fmt.Errorf("failed to repair storage manager. details: %s", err)
	}

	storageManager, err := storage_manager.GetByID(id)
	if err != nil {
		return makeErr(err)
	}

	if storageManager == nil {
		log.Println("storage manager", id, "not found when repairing, assuming it was deleted")
		return nil
	}

	err = k8s_service.RepairStorageManager(id)
	if err != nil {
		return makeErr(err)
	}

	return nil
}
