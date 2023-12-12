package sm_service

import (
	"errors"
	"fmt"
	configModels "go-deploy/models/config"
	"go-deploy/models/sys/storage_manager"
	"go-deploy/pkg/config"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/sm_service/client"
	"go-deploy/service/sm_service/k8s_service"
	"log"
	"sort"
)

// Get gets an existing storage manager.
//
// It supports service.AuthInfo, and will restrict the result to ensure the user has access to the resource.
func (c *Client) Get(id string, opts *client.GetOptions) (*storage_manager.StorageManager, error) {
	sClient := storage_manager.New()

	if c.Auth != nil && !c.Auth.IsAdmin {
		sClient.RestrictToOwner(c.Auth.UserID)
	}

	return c.SM(id, "", sClient)
}

// GetByUserID gets an existing storage by user ID.
//
// It supports service.AuthInfo, and will restrict the result to ensure the user has access to the resource.
func (c *Client) GetByUserID(userID string, opts *client.GetOptions) (*storage_manager.StorageManager, error) {
	sClient := storage_manager.New()

	if c.Auth != nil && userID != c.Auth.UserID && !c.Auth.IsAdmin {
		sClient.RestrictToOwner(c.Auth.UserID)
	} else {
		sClient.RestrictToOwner(userID)
	}

	return c.SM("", userID, sClient)
}

// List lists existing storage managers.
//
// It supports service.AuthInfo, and will restrict the result to ensure the user has access to the resource.
func (c *Client) List(opts *client.ListOptions) ([]storage_manager.StorageManager, error) {
	sClient := storage_manager.New()

	if opts.Pagination != nil {
		sClient.WithPagination(opts.Pagination.Page, opts.Pagination.PageSize)
	}

	if c.Auth != nil && !c.Auth.IsAdmin {
		sClient.RestrictToOwner(c.Auth.UserID)
	}

	resources, err := sClient.List()
	if err != nil {
		return nil, err
	}

	sort.Slice(resources, func(i, j int) bool {
		return resources[i].CreatedAt.Before(resources[j].CreatedAt)
	})

	return resources, nil
}

// Create creates a new storage manager
//
// It returns an error if the storage manager already exists (user ID clash).
func (c *Client) Create(id, userID string, params *storage_manager.CreateParams) error {
	makeErr := func(err error) error {
		return fmt.Errorf("failed to create storage manager. details: %w", err)
	}

	_, err := storage_manager.New().CreateStorageManager(id, userID, params)
	if err != nil {
		if errors.Is(err, storage_manager.AlreadyExistsErr) {
			return sErrors.StorageManagerAlreadyExistsErr
		}

		return makeErr(err)
	}

	err = k8s_service.New(c.Context).Create(id, params)
	if err != nil {
		return makeErr(err)
	}

	return nil
}

// Exists checks if a storage manager exists.
func (c *Client) Exists(userID string) (bool, error) {
	return storage_manager.New().RestrictToOwner(userID).ExistsAny()
}

// Delete deletes an existing storage manager.
//
// It returns an error if the storage manager is not found.
func (c *Client) Delete(id string) error {
	makeErr := func(err error) error {
		return fmt.Errorf("failed to delete storage manager. details: %w", err)
	}

	log.Println("deleting storage manager", id)

	err := k8s_service.New(c.Context).Delete(id)
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

// Repair repairs an existing storage manager.
//
// Trigger repair jobs for every subsystem.
func (c *Client) Repair(id string) error {
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

	err = k8s_service.New(c.Context).Repair(id)
	if err != nil {
		return makeErr(err)
	}

	log.Println("repaired storage manager", id)

	return nil
}

func (c *Client) GetZone() *configModels.DeploymentZone {
	// right now the storage-manager is hosted in se-flem for all users
	zone := "se-flem"

	return config.Config.Deployment.GetZone(zone)
}
