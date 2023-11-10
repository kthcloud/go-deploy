package storage_manager_service

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go-deploy/models/dto/query"
	jobModel "go-deploy/models/sys/job"
	"go-deploy/models/sys/storage_manager"
	"go-deploy/service"
	"go-deploy/service/job_service"
	"go-deploy/service/storage_manager_service/k8s_service"
	"log"
)

var (
	StorageManagerAlreadyExistsErr = fmt.Errorf("storage manager already exists for user")
)

func GetAll(auth *service.AuthInfo) ([]storage_manager.StorageManager, error) {
	if auth.IsAdmin {
		return storage_manager.New().ListAll()
	}

	ownerStorageManager, err := storage_manager.New().RestrictToOwner(auth.UserID).GetOne()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch storage manager. details: %w", err)
	}

	if ownerStorageManager == nil {
		return nil, nil
	}

	return []storage_manager.StorageManager{*ownerStorageManager}, nil
}

func ListAuth(allUsers bool, userID *string, auth *service.AuthInfo, pagination *query.Pagination) ([]storage_manager.StorageManager, error) {
	client := storage_manager.New()

	if pagination != nil {
		client.AddPagination(pagination.Page, pagination.PageSize)
	}

	if userID != nil {
		if *userID != auth.UserID && !auth.IsAdmin {
			return nil, nil
		}
		client.RestrictToOwner(*userID)
	} else if !allUsers || (allUsers && !auth.IsAdmin) {
		client.RestrictToOwner(auth.UserID)
	}

	return client.ListAll()
}

func GetByIdAuth(id string, auth *service.AuthInfo) (*storage_manager.StorageManager, error) {
	storageManager, err := storage_manager.New().GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch storage manager. details: %w", err)
	}

	if storageManager == nil || (storageManager.OwnerID != auth.UserID && !auth.IsAdmin) {
		return nil, nil
	}

	return storageManager, nil
}

func GetByOwnerIdAuth(ownerID string, auth *service.AuthInfo) (*storage_manager.StorageManager, error) {
	storageManager, err := storage_manager.New().RestrictToOwner(ownerID).GetOne()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch storage manager. details: %w", err)
	}

	if storageManager == nil || (storageManager.OwnerID != auth.UserID && !auth.IsAdmin) {
		return nil, nil
	}

	return storageManager, nil
}

func Create(id string, params *storage_manager.CreateParams) error {
	makeErr := func(err error) error {
		return fmt.Errorf("failed to create storage manager. details: %w", err)
	}

	_, err := storage_manager.New().CreateStorageManager(id, params)
	if err != nil {
		if errors.Is(err, storage_manager.AlreadyExistsErr) {
			return StorageManagerAlreadyExistsErr
		}

		return makeErr(err)
	}

	err = k8s_service.Create(id, params)
	if err != nil {
		return makeErr(err)
	}

	return nil
}

func CreateIfNotExists(ownerID string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create storage manager (if not exists). details: %w", err)
	}

	// right now the storage-manager is hosted in se-flem for all users
	zone := "se-flem"

	exists, err := storage_manager.New().RestrictToOwner(ownerID).ExistsAny()
	if err != nil {
		return makeError(err)
	}

	if exists {
		return nil
	}

	storageManagerID := uuid.New().String()
	jobID := uuid.New().String()
	err = job_service.Create(jobID, ownerID, jobModel.TypeCreateStorageManager, map[string]interface{}{
		"id": storageManagerID,
		"params": storage_manager.CreateParams{
			UserID: ownerID,
			Zone:   zone,
		},
	})

	return err
}

func Delete(id string) error {
	makeErr := func(err error) error {
		return fmt.Errorf("failed to delete storage manager. details: %w", err)
	}

	log.Println("deleting storage manager", id)

	err := k8s_service.Delete(id)
	if err != nil {
		return makeErr(err)
	}

	err = storage_manager.New().DeleteByID(id)
	if err != nil {
		return makeErr(err)
	}

	// TODO:
	// New idea! check if everything is deleted at the end, if not, fail the function
	// this will solve the repair <-> delete race, as it will eventually converge as long as the repair
	// respects the beingDeleted activity

	return nil
}

func Repair(id string) error {
	makeErr := func(err error) error {
		return fmt.Errorf("failed to repair storage manager %s. details: %w", id, err)
	}

	storageManager, err := storage_manager.New().GetByID(id)
	if err != nil {
		return makeErr(err)
	}

	if storageManager == nil {
		log.Println("storage manager", id, "not found when repairing, assuming it was deleted")
		return nil
	}

	err = k8s_service.Repair(id)
	if err != nil {
		return makeErr(err)
	}

	return nil
}
