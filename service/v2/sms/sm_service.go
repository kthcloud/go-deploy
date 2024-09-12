package sms

import (
	"errors"
	"fmt"
	configModels "github.com/kthcloud/go-deploy/models/config"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/db/resources/sm_repo"
	"github.com/kthcloud/go-deploy/pkg/log"
	sErrors "github.com/kthcloud/go-deploy/service/errors"
	"github.com/kthcloud/go-deploy/service/utils"
	"github.com/kthcloud/go-deploy/service/v2/sms/k8s_service"
	"github.com/kthcloud/go-deploy/service/v2/sms/opts"
	"sort"
	"strconv"
	"strings"
)

// Get gets an existing storage manager
func (c *Client) Get(id string, opts ...opts.GetOpts) (*model.SM, error) {
	_ = utils.GetFirstOrDefault(opts)

	sClient := sm_repo.New()

	if c.V2.Auth() != nil && !c.V2.Auth().User.IsAdmin {
		sClient.WithOwnerID(c.V2.Auth().User.ID)
	}

	return c.SM(id, "", sClient)
}

// GetByUserID gets an existing storage by user ID
func (c *Client) GetByUserID(userID string, opts ...opts.GetOpts) (*model.SM, error) {
	_ = utils.GetFirstOrDefault(opts)

	sClient := sm_repo.New()

	if c.V2.Auth() != nil && userID != c.V2.Auth().User.ID && !c.V2.Auth().User.IsAdmin {
		sClient.WithOwnerID(c.V2.Auth().User.ID)
	} else {
		sClient.WithOwnerID(userID)
	}

	return c.SM("", userID, sClient)
}

// List lists existing storage managers
func (c *Client) List(opts ...opts.ListOpts) ([]model.SM, error) {
	o := utils.GetFirstOrDefault(opts)

	sClient := sm_repo.New()

	if o.Pagination != nil {
		sClient.WithPagination(o.Pagination.Page, o.Pagination.PageSize)
	}

	if c.V2.Auth() != nil && (!o.All || !c.V2.Auth().User.IsAdmin) {
		sClient.WithOwnerID(c.V2.Auth().User.ID)
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

// Delete deletes an existing storage manager
func (c *Client) Delete(id string) error {
	makeErr := func(err error) error {
		return fmt.Errorf("failed to delete storage manager. details: %w", err)
	}

	log.Println("Deleting storage manager", id)

	err := k8s_service.New(c.Cache).Delete(id)
	if err != nil {
		return makeErr(err)
	}

	return nil
}

// Repair repairs an existing storage manager
func (c *Client) Repair(id string) error {
	makeErr := func(err error) error {
		return fmt.Errorf("failed to repair storage manager %s. details: %w", id, err)
	}

	sm, err := sm_repo.New().GetByID(id)
	if err != nil {
		return makeErr(err)
	}

	if sm == nil {
		log.Println("Storage manager", id, "not found when repairing, assuming it was deleted")
		return nil
	}

	err = k8s_service.New(c.Cache).Repair(id)
	if err != nil {
		return makeErr(err)
	}

	log.Println("Repaired storage manager", id)

	return nil
}

// Exists checks if a storage manager exists
func (c *Client) Exists(userID string) (bool, error) {
	return sm_repo.New().WithOwnerID(userID).ExistsAny()
}

// GetZone returns the deployment zone for the storage manager
func (c *Client) GetZone() *configModels.Zone {
	// Currently, the storage-manager is hosted in the default zone for all users
	zone := config.Config.Deployment.DefaultZone

	return config.Config.GetZone(zone)
}

// GetUrlByUserID returns the URL for the storage manager
func (c *Client) GetUrlByUserID(userID string) *string {
	url, err := sm_repo.New().WithOwnerID(userID).GetURL(getSmExternalPort(c.GetZone().Name))
	if err != nil {
		return nil
	}

	return url
}

func getSmExternalPort(zoneName string) *int {
	zone := config.Config.GetZone(zoneName)
	if zone == nil {
		return nil
	}

	split := strings.Split(zone.Domains.ParentSM, ":")
	if len(split) > 1 {
		port, err := strconv.Atoi(split[1])
		if err != nil {
			return nil
		}

		return &port
	}

	return nil
}
