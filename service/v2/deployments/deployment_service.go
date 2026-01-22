package deployments

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/kthcloud/go-deploy/dto/v2/body"
	configModels "github.com/kthcloud/go-deploy/models/config"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/db/resources/deployment_repo"
	rErrors "github.com/kthcloud/go-deploy/pkg/db/resources/errors"
	"github.com/kthcloud/go-deploy/pkg/db/resources/notification_repo"
	"github.com/kthcloud/go-deploy/pkg/db/resources/resource_migration_repo"
	"github.com/kthcloud/go-deploy/pkg/db/resources/team_repo"
	"github.com/kthcloud/go-deploy/pkg/log"
	sErrors "github.com/kthcloud/go-deploy/service/errors"
	sUtils "github.com/kthcloud/go-deploy/service/utils"
	"github.com/kthcloud/go-deploy/service/v2/deployments/opts"
	"github.com/kthcloud/go-deploy/utils"
	"github.com/kthcloud/go-deploy/utils/subsystemutils"
	"go.mongodb.org/mongo-driver/bson"
)

// Get gets an existing deployment.
//
// It can be fetched in multiple ways including ID, transfer code, and Harbor webhook.
func (c *Client) Get(id string, opts ...opts.GetOpts) (*model.Deployment, error) {
	o := sUtils.GetFirstOrDefault(opts)

	drc := deployment_repo.New()

	if o.MigrationCode != nil {
		rmc := resource_migration_repo.New().
			WithType(model.ResourceMigrationTypeUpdateOwner).
			WithResourceType(model.ResourceTypeDeployment).
			WithCode(*o.MigrationCode)

		migration, err := rmc.Get()
		if err != nil {
			return nil, err
		}

		if migration == nil {
			return nil, nil
		}

		return c.Deployment(migration.ResourceID, drc)
	}

	var effectiveUserID string
	if c.V2.Auth() != nil && !c.V2.Auth().User.IsAdmin {
		effectiveUserID = c.V2.Auth().User.ID
	}

	var teamCheck bool
	if !o.Shared {
		teamCheck = false
	} else if !c.V2.HasAuth() || c.V2.Auth().User.IsAdmin {
		teamCheck = true
	} else {
		var err error
		teamCheck, err = team_repo.New().WithUserID(c.V2.Auth().User.ID).WithResourceID(id).ExistsAny()
		if err != nil {
			return nil, err
		}
	}

	if !teamCheck && effectiveUserID != "" {
		drc.WithOwner(effectiveUserID)
	}

	if o.HarborWebhook != nil {
		return drc.GetByName(o.HarborWebhook.EventData.Repository.Name)
	}

	deployment, err := c.Deployment(id, drc)
	if err != nil {
		return nil, err
	}

	if deployment == nil {
		return nil, nil
	}

	c.markAccessedIfOwner(deployment, drc)

	return deployment, nil
}

// GetByName gets an existing deployment by name.
// This does not support shared deployments.
func (c *Client) GetByName(name string, opts ...opts.GetOpts) (*model.Deployment, error) {
	_ = sUtils.GetFirstOrDefault(opts)

	drc := deployment_repo.New()

	var effectiveUserID string
	if c.V2.Auth() != nil && !c.V2.Auth().User.IsAdmin {
		effectiveUserID = c.V2.Auth().User.ID
	}

	if effectiveUserID != "" {
		drc.WithOwner(effectiveUserID)
	}

	deployment, err := drc.GetByName(name)
	if err != nil {
		return nil, err
	}

	if deployment == nil {
		return nil, nil
	}

	c.markAccessedIfOwner(deployment, drc)

	return deployment, nil
}

// List lists existing deployments.
func (c *Client) List(opts ...opts.ListOpts) ([]model.Deployment, error) {
	o := sUtils.GetFirstOrDefault(opts)

	drc := deployment_repo.New()

	if o.Pagination != nil {
		drc.WithPagination(o.Pagination.Page, o.Pagination.PageSize)
	}

	if o.GpuClaimName != nil {
		if o.GpuClaimRequest != nil {
			drc.WithGpuClaimRequest(*o.GpuClaimName, *o.GpuClaimRequest)
		} else {
			drc.WithGpuClaim(*o.GpuClaimName)
		}
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
		drc.WithOwner(effectiveUserID)
	}

	deployments, err := c.Deployments(drc)
	if err != nil {
		return nil, err
	}

	// Can only view shared if we are listing resources for a specific user
	if o.Shared && effectiveUserID != "" {
		skipIDs := make([]string, len(deployments))
		for i, resource := range deployments {
			skipIDs[i] = resource.ID
		}

		teamClient := team_repo.New().WithUserID(effectiveUserID)
		if o.Pagination != nil {
			teamClient.WithPagination(o.Pagination.Page, o.Pagination.PageSize)
		}

		teams, err := teamClient.List()
		if err != nil {
			return nil, err
		}

		for _, team := range teams {
			for _, resource := range team.GetResourceMap() {
				if resource.Type != model.ResourceTypeDeployment {
					continue
				}

				// Skip existing non-shared resources
				skip := false
				for _, skipID := range skipIDs {
					if resource.ID == skipID {
						skip = true
						break
					}
				}
				if skip {
					continue
				}

				deployment, err := c.Deployment(resource.ID, nil)
				if err != nil {
					return nil, err
				}

				if deployment != nil {
					deployments = append(deployments, *deployment)
				}
			}
		}

		sort.Slice(deployments, func(i, j int) bool {
			return deployments[i].CreatedAt.After(deployments[j].CreatedAt)
		})

		// Since we fetched from two collections, we need to do pagination manually
		if o.Pagination != nil {
			deployments = utils.GetPage(deployments, o.Pagination.PageSize, o.Pagination.Page)
		}

	} else {
		// Sort by createdAt
		sort.Slice(deployments, func(i, j int) bool {
			return deployments[i].CreatedAt.After(deployments[j].CreatedAt)
		})
	}

	for _, deployment := range deployments {
		c.markAccessedIfOwner(&deployment, drc)
	}

	return deployments, nil
}

// Create creates a new deployment.
//
// It returns an error if the deployment already exists (name clash).
func (c *Client) Create(id, ownerID string, deploymentCreate *body.DeploymentCreate) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create deployment. details: %w", err)
	}

	fallbackZone := config.Config.Deployment.DefaultZone
	fallbackImage := createImagePath(ownerID, deploymentCreate.Name)
	fallbackPort := config.Config.Deployment.Port

	params := &model.DeploymentCreateParams{}
	params.FromDTO(deploymentCreate, fallbackZone, fallbackImage, fallbackPort)

	if params.CpuCores == 0 {
		params.CpuCores = config.Config.Deployment.Resources.Limits.CPU
	}

	if params.RAM == 0 {
		params.RAM = config.Config.Deployment.Resources.Limits.RAM
	}

	if !c.V2.System().ZoneHasCapability(params.Zone, configModels.ZoneCapabilityDeployment) {
		return sErrors.NewZoneCapabilityMissingError(params.Zone, configModels.ZoneCapabilityDeployment)
	}

	if len(params.GPUs) > 0 {
		if !c.V2.System().ZoneHasCapability(params.Zone, configModels.ZoneCapabilityDRA) {
			return sErrors.NewZoneCapabilityMissingError(params.Zone, configModels.ZoneCapabilityDRA)
		}

		// TODO: get the roles of the user to verify that the claims exist
		/*var roles = make([]string, 0, 2)

		c.V2.GpuClaims().List(gpuClaimOpts.List{
			Roles:
		})*/
	}

	deployment, err := deployment_repo.New().Create(id, ownerID, params)
	if err != nil {
		if errors.Is(err, rErrors.ErrNonUniqueField) {
			return sErrors.ErrNonUniqueField
		}

		return makeError(err)
	}

	if deployment == nil {
		return makeError(fmt.Errorf("deployment already exists for another user"))
	}

	if deployment.Type == model.DeploymentTypeCustom {
		err = c.Harbor().Create(id, params)
		if err != nil {
			return makeError(err)
		}
	} else {
		err = c.Harbor().CreatePlaceholder(id)
		if err != nil {
			return makeError(err)
		}
	}

	_, err = c.Refresh(id)
	if err != nil {
		return makeError(err)
	}

	err = c.K8s().Create(id, params)
	if err != nil {
		return makeError(err)
	}

	return nil
}

// Update updates an existing deployment.
//
// It returns an error if the deployment is not found.
func (c *Client) Update(id string, dtoUpdate *body.DeploymentUpdate) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update deployment. details: %w", err)
	}

	d, err := c.Get(id, opts.GetOpts{Shared: true})
	if err != nil {
		return makeError(err)
	}

	if d == nil {
		return sErrors.ErrDeploymentNotFound
	}

	mainApp := d.GetMainApp()
	if mainApp == nil {
		return makeError(sErrors.ErrMainAppNotFound)
	}

	params := &model.DeploymentUpdateParams{}
	params.FromDTO(dtoUpdate, d.Type)

	if params.Name != nil && d.Type == model.DeploymentTypeCustom {
		image := createImagePath(d.OwnerID, *params.Name)
		params.Image = &image
	}

	// Don't update the custom domain secret if the update contains the same domain
	if params.CustomDomain != nil && mainApp.CustomDomain != nil && *params.CustomDomain == mainApp.CustomDomain.Domain {
		params.CustomDomain = nil
	}

	err = deployment_repo.New().UpdateWithParams(id, params)
	if err != nil {
		if errors.Is(err, rErrors.ErrNonUniqueField) {
			return sErrors.ErrNonUniqueField
		}

		return makeError(err)
	}

	d, err = c.Refresh(id)
	if err != nil {
		return makeError(err)
	}

	if d.Type == model.DeploymentTypeCustom {
		err = c.Harbor().Update(id, params)
		if err != nil {
			return makeError(err)
		}
	}

	_, err = c.Refresh(id)
	if err != nil {
		return makeError(err)
	}

	err = c.K8s().Update(id, params)
	if err != nil {
		return makeError(err)
	}

	return nil
}

// UpdateOwner updates the owner of the deployment.
//
// This is the second step of the owner update process, where the transfer is actually done.
//
// It returns an error if the deployment is not found.
func (c *Client) UpdateOwner(id string, params *model.DeploymentUpdateOwnerParams) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update deployment owner. details: %w", err)
	}

	d, err := c.Get(id, opts.GetOpts{MigrationCode: params.MigrationCode})
	if err != nil {
		return makeError(err)
	}

	if d == nil {
		return sErrors.ErrDeploymentNotFound
	}

	var newImage *string
	if d.Type == model.DeploymentTypeCustom {
		image := createImagePath(params.NewOwnerID, d.Name)
		newImage = &image
	}

	err = deployment_repo.New().UpdateWithParams(id, &model.DeploymentUpdateParams{
		OwnerID: &params.NewOwnerID,
		Image:   newImage,
	})
	if err != nil {
		return makeError(err)
	}

	_, err = c.Refresh(id)
	if err != nil {
		return makeError(err)
	}

	err = c.Harbor().EnsureOwner(id, params.OldOwnerID)
	if err != nil {
		return makeError(err)
	}

	_, err = c.Refresh(id)
	if err != nil {
		return makeError(err)
	}

	err = c.K8s().EnsureOwner(id, params.OldOwnerID)
	if err != nil {
		return makeError(err)
	}

	err = notification_repo.New().FilterContent("id", id).WithType(model.NotificationResourceTransfer).MarkReadAndCompleted()
	if err != nil {
		return makeError(err)
	}

	log.Println("Deployment", id, "owner updated from", params.OldOwnerID, "to", params.NewOwnerID)
	return nil
}

// Delete deletes an existing deployment.
//
// It returns an error if the deployment is not found.
func (c *Client) Delete(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete deployment. details: %w", err)
	}

	d, err := c.Get(id)
	if err != nil {
		return makeError(err)
	}

	if d == nil {
		return sErrors.ErrDeploymentNotFound
	}

	err = notification_repo.New().FilterContent("id", id).Delete()
	if err != nil {
		return makeError(err)
	}

	err = resource_migration_repo.New().WithResourceID(id).Delete()
	if err != nil {
		return makeError(err)
	}

	err = c.V2.Teams().CleanResource(id)
	if err != nil {
		return makeError(err)
	}

	err = c.Harbor().Delete(id)
	if err != nil {
		return makeError(err)
	}

	err = c.K8s().Delete(id)
	if err != nil {
		return makeError(err)
	}

	return nil
}

// Repair repairs an existing deployment.
//
// Trigger repair jobs for every subsystem.
func (c *Client) Repair(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to repair deployment %s. details: %w", id, err)
	}

	d, err := c.Get(id)
	if err != nil {
		return makeError(err)
	}

	if d == nil {
		return sErrors.ErrDeploymentNotFound
	}

	if !d.Ready() {
		log.Println("Deployment", id, "not ready when repairing.")
		return nil
	}

	// Remove activity if it has been restarting for more than 5 minutes
	now := time.Now()
	if now.Sub(d.RestartedAt) > 5*time.Minute {
		log.Printf("Removing restarting activity from deployment %s\n", d.Name)
		err = deployment_repo.New().RemoveActivity(d.ID, model.ActivityRestarting)
		if err != nil {
			return err
		}
	}

	err = c.K8s().Repair(id)
	if err != nil {
		if errors.Is(err, sErrors.ErrIngressHostInUse) {
			// The user should fix this error, so we don't return an error here
			utils.PrettyPrintError(err)
		} else {
			return makeError(err)
		}
	}

	if !d.Subsystems.Harbor.Placeholder {
		err = c.Harbor().Repair(id)
		if err != nil {
			return makeError(err)
		}
	}

	log.Println("Repaired deployment", id)
	return nil
}

// Restart restarts a deployment.
//
// It is done in best-effort, and only returns an error if any pre-check fails.
func (c *Client) Restart(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to restart deployment. details: %w", err)
	}

	d, err := c.Get(id)
	if err != nil {
		return makeError(err)
	}

	if d == nil {
		return sErrors.ErrDeploymentNotFound
	}

	c.AddLogs(id, model.Log{
		Source: model.LogSourceDeployment,
		Prefix: "[deployment]",
		// Since this is sent as a string, and not a JSON object, we need to prepend the createdAt
		Line:      fmt.Sprintf("%s %s", time.Now().Format(time.RFC3339), "Restart requested"),
		CreatedAt: time.Now(),
	})

	err = deployment_repo.New().SetWithBsonByID(id, bson.D{{Key: "restartedAt", Value: time.Now()}})
	if err != nil {
		return makeError(err)
	}

	err = c.StartActivity(id, model.ActivityRestarting)
	if err != nil {
		return makeError(err)
	}

	go func() {
		// The restart is best-effort, so we mimic a reasonable delay
		time.Sleep(5 * time.Second)

		err = deployment_repo.New().RemoveActivity(id, model.ActivityRestarting)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to remove activity %s from deployment %s. details: %w", model.ActivityRestarting, id, err))
		}
	}()

	err = c.K8s().Restart(id)
	if err != nil {
		return makeError(err)
	}

	return nil
}

// AddLogs adds logs to the deployment.
//
// It is purely done in best-effort
func (c *Client) AddLogs(id string, logs ...model.Log) {
	// logs are added best-effort, so we don't return an error here
	go func() {
		err := deployment_repo.New().AddLogsByName(id, logs...)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to add logs to deployment %s. details: %w", id, err))
		}
	}()
}

// DoCommand executes a command on the deployment.
//
// It is purely done in best-effort
func (c *Client) DoCommand(id string, command string) {
	go func() {
		switch command {
		case "restart":
			err := c.Restart(id)
			if err != nil {
				utils.PrettyPrintError(err)
			}
		}
	}()
}

// CheckQuota checks if the user has enough quota to create or update a deployment.
//
// Make sure to specify either opts.Create or opts.Update in the options (opts.Create takes priority).
//
// It returns an error if the user does not have enough quotas.
func (c *Client) CheckQuota(id string, opts *opts.QuotaOptions) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check quota. details: %w", err)
	}

	if !c.V2.HasAuth() || c.V2.Auth().User.IsAdmin {
		return nil
	}

	usage, err := c.GetUsage(c.V2.Auth().User.ID)
	if err != nil {
		return makeError(err)
	}

	quota := c.V2.Auth().GetEffectiveRole().Quotas

	if opts.Create != nil {
		var replicas int
		var cpu float64
		var ram float64

		if opts.Create.Replicas != nil {
			replicas = *opts.Create.Replicas
		} else {
			replicas = 1
		}

		if opts.Create.CpuCores != nil {
			cpu = usage.CpuCores + *opts.Create.CpuCores*float64(replicas)
		} else {
			cpu = usage.CpuCores + config.Config.Deployment.Resources.Limits.CPU*float64(replicas)
		}

		if opts.Create.RAM != nil {
			ram = usage.RAM + *opts.Create.RAM*float64(replicas)
		} else {
			ram = usage.RAM + config.Config.Deployment.Resources.Limits.RAM*float64(replicas)
		}

		if cpu > quota.CpuCores {
			return sErrors.NewQuotaExceededError(fmt.Sprintf("CPU quota exceeded. Current: %.1f, Quota: %.1f", cpu, quota.CpuCores))
		}

		if ram > quota.RAM {
			return sErrors.NewQuotaExceededError(fmt.Sprintf("RAM quota exceeded. Current: %.1f, Quota: %.1f", ram, quota.RAM))
		}

		return nil
	} else if opts.Update != nil {
		deployment, err := c.Get(id)
		if err != nil {
			return makeError(err)
		}

		if deployment == nil {
			return sErrors.ErrDeploymentNotFound
		}

		replicasBefore := deployment.GetMainApp().Replicas
		cpuBefore := deployment.GetMainApp().CpuCores * float64(replicasBefore)
		ramBefore := deployment.GetMainApp().RAM * float64(replicasBefore)

		var replicasAfter int
		var cpuAfter float64
		var ramAfter float64

		if opts.Update.Replicas != nil {
			replicasAfter = *opts.Update.Replicas
		} else {
			replicasAfter = replicasBefore
		}

		if opts.Update.CpuCores != nil {
			cpuAfter = usage.CpuCores + *opts.Update.CpuCores*float64(replicasAfter) - cpuBefore
		} else {
			cpuAfter = usage.CpuCores + deployment.GetMainApp().CpuCores*float64(replicasAfter) - cpuBefore
		}

		if opts.Update.RAM != nil {
			ramAfter = usage.RAM + *opts.Update.RAM*float64(replicasAfter) - ramBefore
		} else {
			ramAfter = usage.RAM + deployment.GetMainApp().RAM*float64(replicasAfter) - ramBefore
		}

		if cpuAfter > quota.CpuCores {
			return sErrors.NewQuotaExceededError(fmt.Sprintf("CPU quota exceeded. Current: %.1f, Quota: %.1f", cpuAfter, quota.CpuCores))
		}
		if ramAfter > quota.RAM {
			return sErrors.NewQuotaExceededError(fmt.Sprintf("RAM quota exceeded. Current: %.1f, Quota: %.1f", ramAfter, quota.RAM))
		}

		return nil
	} else {
		log.Println("Quota options not set when checking quota for deployment", id)
	}

	return nil
}

// StartActivity starts an activity for the deployment.
//
// It only starts the activity if it is allowed, determined by CanAddActivity.
// It returns a boolean indicating if the activity was started, and a string indicating the reason if it was not.
func (c *Client) StartActivity(id string, activity string) error {
	canAdd, reason := c.CanAddActivity(id, activity)
	if !canAdd {
		if reason == "Deployment not found" {
			return sErrors.ErrDeploymentNotFound
		}

		return sErrors.NewFailedToStartActivityError(reason)
	}

	err := deployment_repo.New().AddActivity(id, activity)
	if err != nil {
		return err
	}

	return nil
}

// CanAddActivity checks if the deployment can add an activity.
//
// It returns a boolean indicating if the activity can be added, and a string indicating the reason if it cannot.
func (c *Client) CanAddActivity(id string, activity string) (bool, string) {
	d, err := c.Deployment(id, nil)
	if err != nil {
		return false, err.Error()
	}

	if d == nil {
		return false, "Deployment not found"
	}

	switch activity {
	case model.ActivityBeingCreated:
		return !d.BeingDeleted(), "Resource is being deleted"
	case model.ActivityBeingDeleted:
		return true, ""
	case model.ActivityUpdating:
		return !d.BeingDeleted() && !d.BeingCreated(), "Resource is being deleted or created"
	case model.ActivityRestarting:
		return !d.BeingDeleted(), "Resource is being deleted"
	case model.ActivityBuilding:
		return !d.BeingDeleted(), "Resource is being deleted"
	case model.ActivityRepairing:
		return d.Ready(), "Resource is not ready"
	}

	return false, fmt.Sprintf("Unknown activity %s", activity)
}

// GetUsage gets the usage of the user.
func (c *Client) GetUsage(userID string) (*model.DeploymentUsage, error) {
	return deployment_repo.New().WithOwner(userID).GetUsage()
}

// NameAvailable checks if a name is available.
func (c *Client) NameAvailable(name string) (bool, error) {
	exists, err := deployment_repo.New().ExistsByName(name)
	if err != nil {
		return false, err
	}

	return !exists, nil
}

// markAccessedIfOwner marks a deployment as accessed if the request is from the owner.
func (c *Client) markAccessedIfOwner(deployment *model.Deployment, drc *deployment_repo.Client) {
	if c.V2.HasAuth() && c.V2.Auth().User.ID == deployment.OwnerID {
		_ = drc.MarkAccessed(deployment.ID)
	}
}

// createImagePath creates a complete container image path that can be pulled from.
func createImagePath(ownerID, name string) string {
	return fmt.Sprintf("%s/%s/%s", config.Config.Registry.URL, subsystemutils.GetPrefixedName(ownerID), name)
}
