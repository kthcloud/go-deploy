package sms

import (
	"errors"
	"fmt"
	configModels "go-deploy/models/config"
	"go-deploy/models/model"
	"go-deploy/pkg/config"
	"go-deploy/pkg/db/resources/sm_repo"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/utils"
	"go-deploy/service/v1/sms/k8s_service"
	"go-deploy/service/v1/sms/opts"
	"log"
	"sort"
)

// Get gets an existing storage manager.
//
// It supports service.AuthInfo, and will restrict the result to ensure the user has access to the model.
func (c *Client) Get(id string, opts ...opts.GetOpts) (*model.SM, error) {
	_ = utils.GetFirstOrDefault(opts)

	sClient := sm_repo.New()

	if c.V1.Auth() != nil && !c.V1.Auth().IsAdmin {
		sClient.RestrictToOwner(c.V1.Auth().UserID)
	}

	return c.SM(id, "", sClient)
}

// GetByUserID gets an existing storage by user ID.
//
// It supports service.AuthInfo, and will restrict the result to ensure the user has access to the model.
func (c *Client) GetByUserID(userID string, opts ...opts.GetOpts) (*model.SM, error) {
	_ = utils.GetFirstOrDefault(opts)

	sClient := sm_repo.New()

	if c.V1.Auth() != nil && userID != c.V1.Auth().UserID && !c.V1.Auth().IsAdmin {
		sClient.RestrictToOwner(c.V1.Auth().UserID)
	} else {
		sClient.RestrictToOwner(userID)
	}

	return c.SM("", userID, sClient)
}

// List lists existing storage managers.
//
// It supports service.AuthInfo, and will restrict the result to ensure the user has access to the model.
func (c *Client) List(opts ...opts.ListOpts) ([]model.SM, error) {
	o := utils.GetFirstOrDefault(opts)

	sClient := sm_repo.New()

	if o.Pagination != nil {
		sClient.WithPagination(o.Pagination.Page, o.Pagination.PageSize)
	}

	if c.V1.Auth() != nil && (!o.All || !c.V1.Auth().IsAdmin) {
		sClient.RestrictToOwner(c.V1.Auth().UserID)
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
func (c *Client) Create(id, userID string, params *model.SmCreateParams) error {
	makeErr := func(err error) error {
		return fmt.Errorf("failed to create storage manager. details: %w", err)
	}

	_, err := sm_repo.New().Create(id, userID, params)
	if err != nil {
		if errors.Is(err, sm_repo.AlreadyExistsErr) {
			return sErrors.SmAlreadyExistsErr
		}

		return makeErr(err)
	}

	err = k8s_service.New(c.Cache).Create(id, params)
	if err != nil {
		return makeErr(err)
	}

	return nil
}

// Delete deletes an existing storage manager.
//
// It returns an error if the storage manager is not found.
func (c *Client) Delete(id string) error {
	makeErr := func(err error) error {
		return fmt.Errorf("failed to delete storage manager. details: %w", err)
	}

	log.Println("deleting storage manager", id)

	err := k8s_service.New(c.Cache).Delete(id)
	if err != nil {
		return makeErr(err)
	}

	return nil
}

// Repair repairs an existing storage manager.
//
// Trigger repair jobs for every subsystem.
func (c *Client) Repair(id string) error {
	makeErr := func(err error) error {
		return fmt.Errorf("failed to repair storage manager %s. details: %w", id, err)
	}

	sm, err := sm_repo.New().GetByID(id)
	if err != nil {
		return makeErr(err)
	}

	if sm == nil {
		log.Println("storage manager", id, "not found when repairing, assuming it was deleted")
		return nil
	}

	err = k8s_service.New(c.Cache).Repair(id)
	if err != nil {
		return makeErr(err)
	}

	log.Println("repaired storage manager", id)

	return nil
}

// Exists checks if a storage manager exists.
func (c *Client) Exists(userID string) (bool, error) {
	return sm_repo.New().RestrictToOwner(userID).ExistsAny()
}

// GetZone returns the deployment zone for the storage manager.
func (c *Client) GetZone() *configModels.DeploymentZone {
	// Currently, the storage-manager is hosted in se-flem for all users
	zone := "se-flem"

	return config.Config.Deployment.GetZone(zone)
}

// GetURL returns the URL for the storage manager.
//
// TODO: Right now it fetches the entire storage manager, but we should probably use a projection.
func (c *Client) GetURL(userID string) *string {
	sm, err := c.GetByUserID(userID)
	if err != nil {
		return nil
	}

	if sm == nil {
		return nil
	}

	return sm.GetURL()
}
