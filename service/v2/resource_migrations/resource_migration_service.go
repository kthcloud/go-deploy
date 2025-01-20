package resource_migrations

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/kthcloud/go-deploy/dto/v2/body"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/models/version"
	"github.com/kthcloud/go-deploy/pkg/db"
	"github.com/kthcloud/go-deploy/pkg/db/resources/deployment_repo"
	"github.com/kthcloud/go-deploy/pkg/db/resources/resource_migration_repo"
	"github.com/kthcloud/go-deploy/pkg/db/resources/vm_repo"
	sErrors "github.com/kthcloud/go-deploy/service/errors"
	sUtils "github.com/kthcloud/go-deploy/service/utils"
	"github.com/kthcloud/go-deploy/service/v2/resource_migrations/opts"
	"github.com/kthcloud/go-deploy/utils"
)

// Get retrieves a resource migration by ID.
func (c *Client) Get(id string, opts ...opts.GetOpts) (*model.ResourceMigration, error) {
	o := sUtils.GetFirstOrDefault(opts)
	rmc := resource_migration_repo.New()

	if o.MigrationCode != nil {
		rmc.WithCode(*o.MigrationCode)
	} else if c.V2.HasAuth() && !c.V2.Auth().User.IsAdmin {
		rmc.WithUserID(c.V2.Auth().User.ID)
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
		if !c.V2.HasAuth() || c.V2.Auth().User.ID == *o.UserID || c.V2.Auth().User.IsAdmin {
			effectiveUserID = *o.UserID
		} else {
			// User cannot access the other user's resources
			effectiveUserID = c.V2.Auth().User.ID
		}
	} else {
		// All deployments are requested
		if c.V2.Auth() != nil && !c.V2.Auth().User.IsAdmin {
			effectiveUserID = c.V2.Auth().User.ID
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

func (c *Client) Create(id, userID string, migrationCreate *body.ResourceMigrationCreate) (*model.ResourceMigration, *string, error) {
	resourceType, err := c.getResourceType(migrationCreate.ResourceID)
	if err != nil {
		return nil, nil, err
	}

	if resourceType == nil {
		return nil, nil, sErrors.ResourceNotFoundErr
	}

	canAccess, err := c.canAccessResource(migrationCreate.ResourceID, *resourceType)
	if err != nil {
		return nil, nil, err
	}

	if !canAccess {
		return nil, nil, sErrors.ResourceNotFoundErr
	}

	switch migrationCreate.Type {
	case model.ResourceMigrationTypeUpdateOwner:
		if migrationCreate.UpdateOwner == nil {
			return nil, nil, sErrors.BadResourceMigrationParamsErr
		}

		ownerID, err := c.getOwnerID(migrationCreate.ResourceID, *resourceType)
		if err != nil {
			return nil, nil, err
		}

		if ownerID == nil {
			return nil, nil, sErrors.ResourceNotFoundErr
		}

		if *ownerID == migrationCreate.UpdateOwner.OwnerID {
			return nil, nil, sErrors.AlreadyMigratedErr
		}

		var status string
		if migrationCreate.Status == nil {
			status = model.ResourceMigrationStatusPending
		} else {
			status = *migrationCreate.Status
		}

		return c.CreateMigrationUpdateOwner(id, userID, migrationCreate.ResourceID, *resourceType, status, &model.ResourceMigrationUpdateOwnerParams{
			NewOwnerID: migrationCreate.UpdateOwner.OwnerID,
			OldOwnerID: *ownerID,
		})
	default:
		return nil, nil, sErrors.BadResourceMigrationTypeErr
	}
}

func (c *Client) CreateMigrationUpdateOwner(id, userID, resourceID, resourceType string, status string, params *model.ResourceMigrationUpdateOwnerParams) (*model.ResourceMigration, *string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create resource migration of type update owner: %w", err)
	}

	// Right now we only support deployments or VMs, so this logic might need to be changed in the future. For teams etc.

	if c.V2.HasAuth() && !c.V2.Auth().User.IsAdmin {
		status = model.ResourceMigrationStatusPending
	}

	code := c.CreateCode()
	rmc := resource_migration_repo.New()
	resourceMigration, err := rmc.Create(id, userID, resourceID, model.ResourceMigrationTypeUpdateOwner, resourceType, &code, status, params)
	if err != nil {
		if errors.Is(err, db.ErrUniqueConstraint) {
			return nil, nil, sErrors.ResourceMigrationAlreadyExistsErr
		}

		return nil, nil, makeError(err)
	}

	switch status {
	case model.ResourceMigrationStatusPending:
		// Create the notification for the user
		name, err := c.getResourceName(resourceID, resourceType)
		if err != nil {
			return nil, nil, makeError(err)
		}

		if name == nil {
			n := "unknown"
			name = &n
		}

		content := map[string]interface{}{
			"id":           resourceMigration.ID,
			"resourceId":   resourceID,
			"resourceName": *name,
			"code":         code,
		}

		if c.V2.HasAuth() {
			content["userId"] = c.V2.Auth().User.ID
		} else {
			user, _ := c.V2.Users().Get(userID)
			if user != nil {
				content["userId"] = user.ID
			}
		}

		_, err = c.V2.Notifications().Create(uuid.NewString(), params.NewOwnerID, &model.NotificationCreateParams{
			Type:    model.NotificationResourceTransfer,
			Content: content,
		})

		if err != nil {
			return nil, nil, makeError(err)
		}

		return resourceMigration, nil, nil
	case model.ResourceMigrationStatusAccepted:
		// Update the owner of the resource
		jobID := uuid.NewString()
		args := map[string]interface{}{
			"id":                  resourceID,
			"resourceMigrationId": resourceMigration.ID,
			"params":              params,
			"authInfo":            c.V2.Auth(),
		}

		switch resourceType {
		case model.ResourceTypeDeployment:
			err = c.V2.Jobs().Create(jobID, userID, model.JobUpdateDeploymentOwner, version.V2, args)
		case model.ResourceTypeVM:
			err = c.V2.Jobs().Create(jobID, userID, model.JobUpdateVmOwner, version.V2, args)
		}
		if err != nil {
			return nil, nil, makeError(err)
		}

		return resourceMigration, &jobID, nil
	default:
		return nil, nil, sErrors.BadResourceMigrationStatusErr
	}
}

func (c *Client) Update(id string, migrationUpdate *body.ResourceMigrationUpdate, opts ...opts.UpdateOpts) (*model.ResourceMigration, *string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update resource migration: %w", err)
	}

	o := sUtils.GetFirstOrDefault(opts)
	rmc := resource_migration_repo.New()

	if o.MigrationCode != nil {
		rmc.WithCode(*o.MigrationCode)
	} else if c.V2.HasAuth() && !c.V2.Auth().User.IsAdmin {
		rmc.WithUserID(c.V2.Auth().User.ID)
	}

	resourceMigration, err := rmc.GetByID(id)
	if err != nil {
		return nil, nil, makeError(err)
	}

	if resourceMigration == nil {
		return nil, nil, sErrors.ResourceMigrationNotFoundErr
	}

	if migrationUpdate.Status == model.ResourceMigrationStatusAccepted {
		if resourceMigration.Status == model.ResourceMigrationStatusAccepted {
			return nil, nil, sErrors.AlreadyAcceptedErr
		}

		canDoUpdate := false

		requireCodeCheck := c.V2.HasAuth() && !c.V2.Auth().User.IsAdmin
		if requireCodeCheck {
			if migrationUpdate.Code == nil {
				return nil, nil, sErrors.BadMigrationCodeErr
			}

			if resourceMigration.Code == nil || *migrationUpdate.Code == *resourceMigration.Code {
				canDoUpdate = true
			}
		} else {
			canDoUpdate = true
		}

		if !canDoUpdate {
			return nil, nil, sErrors.BadMigrationCodeErr
		}

		updateParams := model.ResourceMigrationUpdateParams{}.FromDTO(migrationUpdate)
		err = rmc.UpdateWithParams(id, updateParams)
		if err != nil {
			return nil, nil, makeError(err)
		}

		var jobID *string
		switch resourceMigration.Type {
		case model.ResourceMigrationTypeUpdateOwner:
			jobID, err = c.acceptOwnerUpdate(resourceMigration)
			if err != nil {
				return nil, nil, makeError(err)
			}
		}

		resourceMigration, err = rmc.GetByID(id)
		if err != nil {
			return nil, nil, makeError(err)
		}

		if resourceMigration == nil {
			return nil, nil, sErrors.ResourceMigrationNotFoundErr
		}

		return resourceMigration, jobID, nil
	} else {
		return nil, nil, sErrors.BadResourceMigrationStatusErr
	}
}

func (c *Client) Delete(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete resource migration: %w", err)
	}

	rmc := resource_migration_repo.New()

	if c.V2.HasAuth() && !c.V2.Auth().User.IsAdmin {
		rmc.WithUserID(c.V2.Auth().User.ID)
	}

	err := rmc.DeleteByID(id)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func (c *Client) CreateCode() string {
	return utils.HashStringAlphanumeric(uuid.NewString())
}

func (c *Client) acceptOwnerUpdate(resourceMigration *model.ResourceMigration) (*string, error) {
	if resourceMigration.UpdateOwner == nil {
		return nil, nil
	}

	jobID := uuid.NewString()
	var err error
	switch resourceMigration.ResourceType {
	case model.ResourceTypeDeployment:
		args := map[string]interface{}{
			"id":                  resourceMigration.ResourceID,
			"resourceMigrationId": resourceMigration.ID,
			"params": model.DeploymentUpdateOwnerParams{
				NewOwnerID:    resourceMigration.UpdateOwner.NewOwnerID,
				OldOwnerID:    resourceMigration.UpdateOwner.OldOwnerID,
				MigrationCode: resourceMigration.Code,
			},
			"authInfo": c.V2.Auth(),
		}

		err = c.V2.Jobs().Create(jobID, resourceMigration.UserID, model.JobUpdateDeploymentOwner, version.V2, args)
	case model.ResourceTypeVM:
		args := map[string]interface{}{
			"id":                  resourceMigration.ResourceID,
			"resourceMigrationId": resourceMigration.ID,
			"params": model.VmUpdateOwnerParams{
				NewOwnerID:    resourceMigration.UpdateOwner.NewOwnerID,
				OldOwnerID:    resourceMigration.UpdateOwner.OldOwnerID,
				MigrationCode: resourceMigration.Code,
			},
			"authInfo": c.V2.Auth(),
		}

		err = c.V2.Jobs().Create(jobID, resourceMigration.UserID, model.JobUpdateVmOwner, version.V2, args)
	}
	if err != nil {
		return nil, err
	}

	return &jobID, nil
}

func (c *Client) canAccessResource(id string, resourceType string) (bool, error) {
	if !c.V2.HasAuth() {
		return true, nil
	}

	switch resourceType {
	case model.ResourceTypeVM:
		vm, err := c.V2.VMs().Get(id)
		if err != nil {
			return false, err
		}

		if vm == nil {
			return false, nil
		}
	case model.ResourceTypeDeployment:
		deployment, err := c.V2.Deployments().Get(id)
		if err != nil {
			return false, err
		}

		if deployment == nil {
			return false, nil
		}
	}

	return true, nil
}

func (c *Client) getResourceName(id string, resourceType string) (*string, error) {
	switch resourceType {
	case model.ResourceTypeVM:
		return vm_repo.New(version.V2).GetName(id)
	case model.ResourceTypeDeployment:
		return deployment_repo.New().GetName(id)
	default:
		return nil, sErrors.BadResourceMigrationResourceTypeErr
	}
}

func (c *Client) getResourceType(id string) (*string, error) {
	exists, err := vm_repo.New(version.V2).ExistsByID(id)
	if err != nil {
		return nil, err
	}

	if exists {
		t := model.ResourceTypeVM
		return &t, nil
	}

	exists, err = deployment_repo.New().ExistsByID(id)
	if err != nil {
		return nil, err
	}

	if exists {
		t := model.ResourceTypeDeployment
		return &t, nil
	}

	return nil, nil
}

func (c *Client) getOwnerID(id string, resourceType string) (*string, error) {
	switch resourceType {
	case model.ResourceTypeVM:
		vm, err := vm_repo.New(version.V2).GetByID(id)
		if err != nil {
			return nil, err
		}

		if vm == nil {
			return nil, sErrors.ResourceNotFoundErr
		}

		return &vm.OwnerID, nil
	case model.ResourceTypeDeployment:
		deployment, err := deployment_repo.New().GetByID(id)
		if err != nil {
			return nil, err
		}

		if deployment == nil {
			return nil, sErrors.ResourceNotFoundErr
		}

		return &deployment.OwnerID, nil
	}

	return nil, sErrors.BadResourceMigrationResourceTypeErr
}
