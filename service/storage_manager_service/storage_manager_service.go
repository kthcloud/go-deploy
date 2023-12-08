package storage_manager_service

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	deploymentModel "go-deploy/models/sys/deployment"
	jobModel "go-deploy/models/sys/job"
	"go-deploy/models/sys/storage_manager"
	"go-deploy/service/job_service"
	"go-deploy/service/storage_manager_service/client"
	"go-deploy/service/storage_manager_service/k8s_service"
	"log"
)

var (
	StorageManagerAlreadyExistsErr = fmt.Errorf("storage manager already exists for user")
)

func (c *Client) Get(opts *client.GetOptions) (*storage_manager.StorageManager, error) {
	dClient := deploymentModel.New()

	var effectiveUserID string
	if !c.HasUserID() {
		if c.Auth == nil || c.Auth.UserID == c.UserID() || c.Auth.IsAdmin {
			effectiveUserID = c.UserID()
		} else {
			effectiveUserID = c.Auth.UserID
		}
	} else {
		if c.Auth != nil && !c.Auth.IsAdmin {
			effectiveUserID = c.Auth.UserID
		}
	}

	if effectiveUserID != "" {
		dClient.RestrictToOwner(effectiveUserID)
	}

	return c.StorageManager(), nil
}

func (c *Client) List(opts *client.ListOptions) ([]storage_manager.StorageManager, error) {
	sClient := storage_manager.New()

	if opts.Pagination != nil {
		sClient.WithPagination(opts.Pagination.Page, opts.Pagination.PageSize)
	}

	var effectiveUserID string
	if c.Auth == nil || c.Auth.IsAdmin {
		effectiveUserID = c.UserID()
	} else {
		effectiveUserID = c.Auth.UserID
	}

	if effectiveUserID != "" {
		sClient.RestrictToOwner(effectiveUserID)
	}

	resources, err := sClient.List()
	if err != nil {
		return nil, err
	}

	return resources, nil
}

func (c *Client) Create(params *storage_manager.CreateParams) error {
	makeErr := func(err error) error {
		return fmt.Errorf("failed to create storage manager. details: %w", err)
	}

	if !c.HasID() || !c.HasUserID() {
		return errors.New("invalid create request. missing id or user id")
	}

	_, err := storage_manager.New().CreateStorageManager(c.ID(), c.UserID(), params)
	if err != nil {
		if errors.Is(err, storage_manager.AlreadyExistsErr) {
			return StorageManagerAlreadyExistsErr
		}

		return makeErr(err)
	}

	err = k8s_service.New(&c.Context).Create(params)
	if err != nil {
		return makeErr(err)
	}

	return nil
}

func (c *Client) CreateIfNotExists() error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create storage manager (if not exists). details: %w", err)
	}

	if !c.HasUserID() {
		return errors.New("invalid create if not exists request. missing user id")
	}

	// right now the storage-manager is hosted in se-flem for all users
	zone := "se-flem"

	exists, err := storage_manager.New().RestrictToOwner(c.UserID()).ExistsAny()
	if err != nil {
		return makeError(err)
	}

	if exists {
		return nil
	}

	storageManagerID := uuid.New().String()
	jobID := uuid.New().String()
	err = job_service.Create(jobID, c.UserID(), jobModel.TypeCreateStorageManager, map[string]interface{}{
		"id":     storageManagerID,
		"userId": c.UserID(),
		"params": storage_manager.CreateParams{
			Zone: zone,
		},
	})

	return err
}

func (c *Client) Delete() error {
	makeErr := func(err error) error {
		return fmt.Errorf("failed to delete storage manager. details: %w", err)
	}

	log.Println("deleting storage manager", c.ID())

	err := k8s_service.New(&c.Context).Delete()
	if err != nil {
		return makeErr(err)
	}

	err = storage_manager.New().DeleteByID(c.ID())
	if err != nil {
		return makeErr(err)
	}

	// TODO:
	// New idea! check if everything is deleted at the end, if not, fail the function
	// this will solve the repair <-> delete race, as it will eventually converge as long as the repair
	// respects the beingDeleted activity

	return nil
}

func (c *Client) Repair() error {
	makeErr := func(err error) error {
		return fmt.Errorf("failed to repair storage manager %s. details: %w", c.ID(), err)
	}

	storageManager, err := storage_manager.New().GetByID(c.ID())
	if err != nil {
		return makeErr(err)
	}

	if storageManager == nil {
		log.Println("storage manager", c.ID(), "not found when repairing, assuming it was deleted")
		return nil
	}

	err = k8s_service.New(&c.Context).Repair()
	if err != nil {
		return makeErr(err)
	}

	log.Println("repaired storage manager", c.ID())

	return nil
}
