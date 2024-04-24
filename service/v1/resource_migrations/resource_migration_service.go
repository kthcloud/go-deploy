package resource_migrations

import (
	"fmt"
	"github.com/google/uuid"
	"go-deploy/dto/v1/body"
	"go-deploy/models/model"
	"go-deploy/models/version"
	"go-deploy/pkg/db/resources/deployment_repo"
	"go-deploy/pkg/db/resources/resource_migration_repo"
	"go-deploy/pkg/db/resources/vm_repo"
	sErrors "go-deploy/service/errors"
	sUtils "go-deploy/service/utils"
	"go-deploy/service/v1/resource_migrations/opts"
	"go-deploy/utils"
)

// Get retrieves a resource migration by ID.
func (c *Client) Get(id string, opts ...opts.GetOpts) (*model.ResourceMigration, error) {
	rmc := resource_migration_repo.New()

	if c.V1.HasAuth() && !c.V1.Auth().IsAdmin {
		rmc.WithUserID(c.V1.Auth().UserID)
	}

	resourceMigration, err := rmc.GetByID(id)
	if err != nil {
		return nil, err
	}

	return resourceMigration, nil
}

// List retrieves a list of resource migrations.
func (c *Client) List(opts ...opts.ListOpts) ([]model.ResourceMigration, error) {
	o := sUtils.GetFirstOrDefault(opts)

	rmcClient := resource_migration_repo.New()

	if o.Pagination != nil {
		rmcClient.WithPagination(o.Pagination.Page, o.Pagination.PageSize)
	}

	var effectiveUserID string
	if o.UserID != nil {
		// Specific user's deployments are requested
		if !c.V1.HasAuth() || c.V1.Auth().UserID == *o.UserID || c.V1.Auth().IsAdmin {
			effectiveUserID = *o.UserID
		} else {
			// User cannot access the other user's resources
			effectiveUserID = c.V1.Auth().UserID
		}
	} else {
		// All deployments are requested
		if c.V1.Auth() != nil && !c.V1.Auth().IsAdmin {
			effectiveUserID = c.V1.Auth().UserID
		}
	}

	if effectiveUserID != "" {
		rmcClient.WithUserID(effectiveUserID)
	}

	resourceMigrations, err := rmcClient.List()
	if err != nil {
		return nil, err
	}

	return resourceMigrations, nil
}

func (c *Client) Create(id, userID string, migrationCreate *body.ResourceMigrationCreate) error {
	switch migrationCreate.Type {
	case model.ResourceMigrationTypeUpdateOwner:
		if migrationCreate.UpdateOwner == nil {
			return sErrors.BadResourceMigrationParamsErr
		}

		return c.CreateUpdateOwnerMigration(id, userID, migrationCreate.ResourceID, migrationCreate.ResourceType, &model.ResourceMigrationUpdateOwnerParams{
			OwnerID: migrationCreate.UpdateOwner.OwnerID,
		})
	default:
		return sErrors.BadResourceMigrationTypeErr
	}
}

func (c *Client) CreateUpdateOwnerMigration(id, userID, resourceID, resourceType string, params *model.ResourceMigrationUpdateOwnerParams) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create resource migration of type update owner: %w", err)
	}

	// Right now we only support deployments or VMs, so this logic might need to be changed in the future.
	// The flow is the same for both: If admin, set status to ready, otherwise set to pending.
	// This is because we don't want to allow users to update the owner of a resource without approval.

	var status string
	if c.V1.HasAuth() && c.V1.Auth().IsAdmin {
		status = model.ResourceMigrationStatusReady
	} else {
		status = model.ResourceMigrationStatusPending
	}

	rmc := resource_migration_repo.New()
	_, err := rmc.Create(id, userID, resourceID, model.ResourceMigrationTypeUpdateOwner, resourceType, status, params)
	if err != nil {
		return makeError(err)
	}

	// Create the notification for the user
	name, err := c.getResourceName(resourceID, resourceType)
	if err != nil {
		return makeError(err)
	}

	if name == nil {
		n := "unknown"
		name = &n
	}

	content := map[string]interface{}{
		"id":   resourceID,
		"name": *name,
		"code": c.CreateTransferCode(),
	}

	if c.V1.HasAuth() {
		content["user"] = c.V1.Auth().UserID
		content["email"] = c.V1.Auth().GetEmail()
	}

	_, err = c.V1.Notifications().Create(uuid.NewString(), params.OwnerID, &model.NotificationCreateParams{
		Type:    model.NotificationDeploymentTransfer,
		Content: content,
	})

	if err != nil {
		return makeError(err)
	}

	return nil
}

func (c *Client) Update(id string, migrationUpdate *body.ResourceMigrationUpdate) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update resource migration: %w", err)
	}

	rmc := resource_migration_repo.New()

	if c.V1.HasAuth() && !c.V1.Auth().IsAdmin {
		rmc.WithUserID(c.V1.Auth().UserID)
	}

	resourceMigration, err := rmc.GetByID(id)
	if err != nil {
		return makeError(err)
	}

	if resourceMigration == nil {
		return sErrors.ResourceMigrationNotFoundErr
	}

	if migrationUpdate.Status != model.ResourceMigrationStatusReady {
		return sErrors.BadResourceMigrationStatusErr
	}

	canDoUpdate := false

	switch resourceMigration.Type {
	case model.ResourceMigrationTypeUpdateOwner:
		requireTokenCheck := c.V1.HasAuth() && !c.V1.Auth().IsAdmin
		if requireTokenCheck {
			if migrationUpdate.Token == nil {
				return sErrors.BadResourceMigrationParamsErr
			}

			if resourceMigration.Token == nil || *migrationUpdate.Token == *resourceMigration.Token {
				canDoUpdate = true
			}
		}

	default:
		return sErrors.BadResourceMigrationTypeErr
	}

	if canDoUpdate {
		updateParams := model.ResourceMigrationUpdateParams{}.FromDTO(migrationUpdate)

		err = rmc.UpdateWithParams(id, updateParams)
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

func (c *Client) Delete(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete resource migration: %w", err)
	}

	rmc := resource_migration_repo.New()

	if c.V1.HasAuth() && !c.V1.Auth().IsAdmin {
		rmc.WithUserID(c.V1.Auth().UserID)
	}

	err := rmc.DeleteByID(id)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func (c *Client) CreateTransferCode() string {
	return utils.HashStringAlphanumeric(uuid.NewString())
}

func (c *Client) getResourceName(id string, resourceType string) (*string, error) {
	switch resourceType {
	case model.ResourceMigrationResourceTypeVM:
		return vm_repo.New(version.V2).GetName(id)
	case model.ResourceMigrationResourceTypeDeployment:
		return deployment_repo.New().GetName(id)
	default:
		return nil, sErrors.BadResourceMigrationResourceTypeErr
	}
}
