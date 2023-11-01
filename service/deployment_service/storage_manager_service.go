package deployment_service

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go-deploy/models/dto/query"
	storageManagerModel "go-deploy/models/sys/deployment/storage_manager"
	jobModel "go-deploy/models/sys/job"
	"go-deploy/service"
	"go-deploy/service/deployment_service/k8s_service"
	"go-deploy/service/job_service"
	"log"
)

var (
	StorageManagerAlreadyExistsErr = fmt.Errorf("storage manager already exists for user")
)

func GetAllStorageManagers(auth *service.AuthInfo) ([]storageManagerModel.StorageManager, error) {
	if auth.IsAdmin {
		return storageManagerModel.New().ListAll()
	}

	ownerStorageManager, err := storageManagerModel.New().RestrictToOwner(auth.UserID).GetOne()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch storage manager. details: %w", err)
	}

	if ownerStorageManager == nil {
		return nil, nil
	}

	return []storageManagerModel.StorageManager{*ownerStorageManager}, nil
}

func ListStorageManagersAuth(allUsers bool, userID *string, auth *service.AuthInfo, pagination *query.Pagination) ([]storageManagerModel.StorageManager, error) {
	client := storageManagerModel.New()

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

func GetStorageManagerByIdAuth(id string, auth *service.AuthInfo) (*storageManagerModel.StorageManager, error) {
	storageManager, err := storageManagerModel.New().GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch storage manager. details: %w", err)
	}

	if storageManager == nil || (storageManager.OwnerID != auth.UserID && !auth.IsAdmin) {
		return nil, nil
	}

	return storageManager, nil
}

func GetStorageManagerByOwnerIdAuth(ownerID string, auth *service.AuthInfo) (*storageManagerModel.StorageManager, error) {
	storageManager, err := storageManagerModel.New().RestrictToOwner(ownerID).GetOne()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch storage manager. details: %w", err)
	}

	if storageManager == nil || (storageManager.OwnerID != auth.UserID && !auth.IsAdmin) {
		return nil, nil
	}

	return storageManager, nil
}

func CreateStorageManager(id string, params *storageManagerModel.CreateParams) error {
	makeErr := func(err error) error {
		return fmt.Errorf("failed to create storage manager. details: %w", err)
	}

	_, err := storageManagerModel.New().CreateStorageManager(id, params)
	if err != nil {
		if errors.Is(err, storageManagerModel.AlreadyExistsErr) {
			return StorageManagerAlreadyExistsErr
		}

		return makeErr(err)
	}

	err = k8s_service.CreateStorageManager(id, params)
	if err != nil {
		return makeErr(err)
	}

	return nil
}

func CreateStorageManagerIfNotExists(ownerID string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create storage manager (if not exists). details: %w", err)
	}

	// right now the storage-manager is hosted in se-flem for all users
	zone := "se-flem"

	exists, err := storageManagerModel.New().RestrictToOwner(ownerID).ExistsAny()
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
		"params": storageManagerModel.CreateParams{
			UserID: ownerID,
			Zone:   zone,
		},
	})

	return err
}

func DeleteStorageManager(id string) error {
	makeErr := func(err error) error {
		return fmt.Errorf("failed to delete storage manager. details: %w", err)
	}

	log.Println("deleting storage manager", id)

	err := k8s_service.DeleteStorageManager(id)
	if err != nil {
		return makeErr(err)
	}

	err = storageManagerModel.New().DeleteByID(id)
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
		return fmt.Errorf("failed to repair storage manager. details: %w", err)
	}

	storageManager, err := storageManagerModel.New().GetByID(id)
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
