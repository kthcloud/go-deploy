package deployments

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go-deploy/dto/v1/body"
	"go-deploy/models/model"
	"go-deploy/models/version"
	"go-deploy/pkg/config"
	"go-deploy/pkg/db/resources/deployment_repo"
	"go-deploy/pkg/db/resources/notification_repo"
	"go-deploy/pkg/db/resources/team_repo"
	sErrors "go-deploy/service/errors"
	utils2 "go-deploy/service/utils"
	"go-deploy/service/v1/deployments/opts"
	"go-deploy/utils"
	"go-deploy/utils/subsystemutils"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"sort"
	"time"
)

// Get gets an existing deployment.
//
// It can be fetched in multiple ways including ID, name, transfer code, and Harbor webhook.
// It supports service.AuthInfo, and will restrict the result to ensure the user has access to the model.
func (c *Client) Get(id string, opts ...opts.GetOpts) (*model.Deployment, error) {
	o := utils2.GetFirstOrDefault(opts)

	dClient := deployment_repo.New()

	if o.TransferCode != "" {
		return dClient.WithTransferCode(o.TransferCode).Get()
	}

	var effectiveUserID string
	if c.V1.Auth() != nil && !c.V1.Auth().IsAdmin {
		effectiveUserID = c.V1.Auth().UserID
	}

	var teamCheck bool
	if !o.Shared {
		teamCheck = false
	} else if !c.V1.HasAuth() || c.V1.Auth().IsAdmin {
		teamCheck = true
	} else {
		var err error
		teamCheck, err = team_repo.New().WithUserID(c.V1.Auth().UserID).WithResourceID(id).ExistsAny()
		if err != nil {
			return nil, err
		}
	}

	if !teamCheck && effectiveUserID != "" {
		dClient.WithOwner(effectiveUserID)
	}

	if o.HarborWebhook != nil {
		return dClient.GetByName(o.HarborWebhook.EventData.Repository.Name)
	}

	return c.Deployment(id, dClient)
}

// List lists existing deployments.
//
// It supports service.AuthInfo, and will restrict the result to ensure the user has access to the model.
func (c *Client) List(opts ...opts.ListOpts) ([]model.Deployment, error) {
	o := utils2.GetFirstOrDefault(opts)

	dClient := deployment_repo.New()

	if o.Pagination != nil {
		dClient.WithPagination(o.Pagination.Page, o.Pagination.PageSize)
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
		dClient.WithOwner(effectiveUserID)
	}

	resources, err := c.Deployments(dClient)
	if err != nil {
		return nil, err
	}

	// Can only view shared if we are listing resources for a specific user
	if o.Shared && effectiveUserID != "" {
		skipIDs := make([]string, len(resources))
		for i, resource := range resources {
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
				if resource.Type != model.TeamResourceDeployment {
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
					resources = append(resources, *deployment)
				}
			}
		}

		sort.Slice(resources, func(i, j int) bool {
			return resources[i].CreatedAt.After(resources[j].CreatedAt)
		})

		// Since we fetched from two collections, we need to do pagination manually
		if o.Pagination != nil {
			resources = utils.GetPage(resources, o.Pagination.PageSize, o.Pagination.Page)
		}

	} else {
		// Sort by createdAt
		sort.Slice(resources, func(i, j int) bool {
			return resources[i].CreatedAt.After(resources[j].CreatedAt)
		})
	}

	return resources, nil
}

// Create creates a new deployment.
//
// It returns an error if the deployment already exists (name clash).
//
// If GitHub is requested, it will also manually trigger a build to the latest commit.
func (c *Client) Create(id, ownerID string, deploymentCreate *body.DeploymentCreate) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create deployment. details: %w", err)
	}

	// temporary hard-coded fallback
	fallbackZone := "se-flem"
	fallbackImage := createImagePath(ownerID, deploymentCreate.Name)
	fallbackPort := config.Config.Deployment.Port

	params := &model.DeploymentCreateParams{}
	params.FromDTO(deploymentCreate, fallbackZone, fallbackImage, fallbackPort)

	d, err := deployment_repo.New().Create(id, ownerID, params)
	if err != nil {
		if errors.Is(err, deployment_repo.NonUniqueFieldErr) {
			return sErrors.NonUniqueFieldErr
		}

		return makeError(err)
	}

	if d == nil {
		return makeError(fmt.Errorf("deployment already exists for another user"))
	}

	if d.Type == model.DeploymentTypeCustom {
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

	d, err = c.Refresh(id)
	if err != nil {
		return makeError(err)
	}

	err = c.K8s().Create(id, params)
	if err != nil {
		return makeError(err)
	}

	return nil
}

// kubernetes.io/metadata.name
// owner-id: 955f0f87-37fd-4792-90eb-9bf6989e698a

// Update updates an existing deployment.
//
// It returns an error if the deployment is not found.
func (c *Client) Update(id string, dtoUpdate *body.DeploymentUpdate) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update deployment. details: %w", err)
	}

	d, err := c.Deployment(id, nil)
	if err != nil {
		return makeError(err)
	}

	if d == nil {
		return sErrors.DeploymentNotFoundErr
	}

	mainApp := d.GetMainApp()
	if mainApp == nil {
		return makeError(sErrors.MainAppNotFoundErr)
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
		if errors.Is(err, deployment_repo.NonUniqueFieldErr) {
			return sErrors.NonUniqueFieldErr
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

	err = c.K8s().Update(id, params)
	if err != nil {
		return makeError(err)
	}

	return nil
}

// UpdateOwnerSetup updates the owner of the deployment.
//
// This is the first step of the owner update process, where it is decided if a notification should be created,
// or if the transfer should be done immediately.
//
// It returns an error if the deployment is not found.
func (c *Client) UpdateOwnerSetup(id string, params *body.DeploymentUpdateOwner) (*string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update deployment owner. details: %w", err)
	}

	d, err := c.Deployment(id, nil)
	if err != nil {
		return nil, makeError(err)
	}

	if d == nil {
		return nil, sErrors.DeploymentNotFoundErr
	}

	if d.OwnerID == params.NewOwnerID {
		return nil, nil
	}

	doTransfer := false

	if !c.V1.HasAuth() || c.V1.Auth().IsAdmin {
		doTransfer = true
	} else if c.V1.Auth().UserID == params.NewOwnerID {
		if params.TransferCode == nil || d.Transfer == nil || d.Transfer.Code != *params.TransferCode {
			return nil, sErrors.InvalidTransferCodeErr
		}

		doTransfer = true
	}

	var effectiveUserID string
	if !c.V1.HasAuth() {
		effectiveUserID = "system"
	} else {
		effectiveUserID = c.V1.Auth().UserID
	}

	if doTransfer {
		jobID := uuid.New().String()
		err := c.V1.Jobs().Create(jobID, effectiveUserID, model.JobUpdateDeploymentOwner, version.V1, map[string]interface{}{
			"id":     id,
			"params": *params,
		})

		if err != nil {
			return nil, makeError(err)
		}

		return &jobID, nil
	}

	// Create a transfer notification
	code := createTransferCode()
	err = deployment_repo.New().UpdateWithParams(id, &model.DeploymentUpdateParams{
		TransferUserID: &params.NewOwnerID,
		TransferCode:   &code,
	})
	if err != nil {
		return nil, makeError(err)
	}

	_, err = c.V1.Notifications().Create(uuid.NewString(), params.NewOwnerID, &model.NotificationCreateParams{
		Type: model.NotificationDeploymentTransfer,
		Content: map[string]interface{}{
			"id":     d.ID,
			"name":   d.Name,
			"userId": params.OldOwnerID,
			"email":  c.V1.Auth().GetEmail(),
			"code":   code,
		},
	})

	if err != nil {
		return nil, makeError(err)
	}

	return nil, nil
}

// UpdateOwner updates the owner of the deployment.
//
// This is the second step of the owner update process, where the transfer is actually done.
//
// It returns an error if the deployment is not found.
func (c *Client) UpdateOwner(id string, params *body.DeploymentUpdateOwner) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update deployment owner. details: %w", err)
	}

	d, err := c.Deployment(id, nil)
	if err != nil {
		return makeError(err)
	}

	if d == nil {
		return sErrors.DeploymentNotFoundErr
	}

	var newImage *string
	if d.Type == model.DeploymentTypeCustom {
		image := createImagePath(params.NewOwnerID, d.Name)
		newImage = &image
	}

	emptyString := ""

	err = deployment_repo.New().UpdateWithParams(id, &model.DeploymentUpdateParams{
		OwnerID:        &params.NewOwnerID,
		Image:          newImage,
		TransferCode:   &emptyString,
		TransferUserID: &emptyString,
	})
	if err != nil {
		return makeError(err)
	}

	err = c.Harbor().EnsureOwner(id, params.OldOwnerID)
	if err != nil {
		return makeError(err)
	}

	err = c.K8s().EnsureOwner(id, params.OldOwnerID)
	if err != nil {
		return makeError(err)
	}

	nmc := notification_repo.New().WithUserID(params.NewOwnerID).FilterContent("id", id).WithType(model.NotificationDeploymentTransfer)
	err = nmc.MarkReadAndCompleted()
	if err != nil {
		return makeError(err)
	}

	log.Println("deployment", id, "owner updated from", params.OldOwnerID, "to", params.NewOwnerID)
	return nil
}

// ClearUpdateOwner clears the owner update process.
//
// This is intended to be used when the owner update process is canceled.
func (c *Client) ClearUpdateOwner(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to clear deployment owner update. details: %w", err)
	}

	d, err := c.Deployment(id, nil)
	if err != nil {
		return makeError(err)
	}

	if d == nil {
		return sErrors.DeploymentNotFoundErr
	}

	if d.Transfer == nil {
		return nil
	}

	emptyString := ""
	err = deployment_repo.New().UpdateWithParams(id, &model.DeploymentUpdateParams{
		TransferUserID: &emptyString,
		TransferCode:   &emptyString,
	})
	if err != nil {
		return makeError(err)
	}

	// TODO: delete notification?

	return nil
}

// Delete deletes an existing deployment.
//
// It returns an error if the deployment is not found.
func (c *Client) Delete(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete deployment. details: %w", err)
	}

	d, err := c.Deployment(id, nil)
	if err != nil {
		return makeError(err)
	}

	if d == nil {
		return sErrors.DeploymentNotFoundErr
	}

	nmc := notification_repo.New().FilterContent("id", id)
	err = nmc.Delete()
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

	d, err := c.Deployment(id, nil)
	if err != nil {
		return makeError(err)
	}

	if d == nil {
		return sErrors.DeploymentNotFoundErr
	}

	if !d.Ready() {
		log.Println("deployment", id, "not ready when repairing.")
		return nil
	}

	err = c.K8s().Repair(id)
	if err != nil {
		if errors.Is(err, sErrors.IngressHostInUseErr) {
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

	log.Println("repaired deployment", id)
	return nil
}

// Restart restarts a deployment.
//
// It is done in best-effort, and only returns an error if any pre-check fails.
func (c *Client) Restart(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to restart deployment. details: %w", err)
	}

	d, err := c.Deployment(id, nil)
	if err != nil {
		return makeError(err)
	}

	if d == nil {
		return sErrors.DeploymentNotFoundErr
	}

	c.AddLogs(id, model.Log{
		Source: model.LogSourceDeployment,
		Prefix: "[deployment]",
		// Since this is sent as a string, and not a JSON object, we need to prepend the createdAt
		Line:      fmt.Sprintf("%s %s", time.Now().Format(time.RFC3339), "Restart requested"),
		CreatedAt: time.Now(),
	})

	err = deployment_repo.New().SetWithBsonByID(id, bson.D{{"restartedAt", time.Now()}})
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

	if !c.V1.HasAuth() || c.V1.Auth().IsAdmin {
		return nil
	}

	usage, err := c.GetUsage(c.V1.Auth().UserID)
	if err != nil {
		return makeError(err)
	}

	quota := c.V1.Auth().GetEffectiveRole().Quotas

	if opts.Create != nil {
		add := 1
		if opts.Create.Replicas != nil {
			add = *opts.Create.Replicas
		}

		// Ensure that users do not create infinite deployments with 0 replicas
		if add == 0 {
			add = 1
		}

		totalCount := usage.Count + add

		if totalCount > quota.Deployments {
			return sErrors.NewQuotaExceededError(fmt.Sprintf("Deployment quota exceeded. Current: %d, Quota: %d", totalCount, quota.Deployments))
		}

		return nil
	} else if opts.Update != nil {
		d, err := c.Deployment(id, nil)
		if err != nil {
			return makeError(err)
		}

		if d == nil {
			return sErrors.DeploymentNotFoundErr
		}

		if opts.Update.Replicas != nil {
			totalBefore := usage.Count
			replicasReq := *opts.Update.Replicas

			// Ensure that users do not create infinite deployments with 0 replicas
			if replicasReq == 0 {
				replicasReq = 1
			}

			add := replicasReq - d.GetMainApp().Replicas

			totalAfter := totalBefore + add

			if totalAfter > quota.Deployments {
				return sErrors.NewQuotaExceededError(fmt.Sprintf("Deployment quota exceeded. Current: %d, Quota: %d", totalAfter, quota.Deployments))
			}
		}

		return nil
	} else {
		log.Println("quota options not set when checking quota for deployment", id)
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
			return sErrors.DeploymentNotFoundErr
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
		return !d.BeingDeleted(), "TeamResource is being deleted"
	case model.ActivityBeingDeleted:
		return true, ""
	case model.ActivityUpdating:
		return !d.BeingDeleted() && !d.BeingCreated(), "TeamResource is being deleted or created"
	case model.ActivityRestarting:
		return !d.BeingDeleted(), "TeamResource is being deleted"
	case model.ActivityBuilding:
		return !d.BeingDeleted(), "TeamResource is being deleted"
	case model.ActivityRepairing:
		return d.Ready(), "TeamResource is not ready"
	}

	return false, fmt.Sprintf("Unknown activity %s", activity)
}

// GetUsage gets the usage of the user.
func (c *Client) GetUsage(userID string) (*model.DeploymentUsage, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get usage. details: %w", err)
	}

	count, err := deployment_repo.New().WithOwner(userID).CountReplicas()
	if err != nil {
		return nil, makeError(err)
	}

	return &model.DeploymentUsage{
		Count: count,
	}, nil
}

// NameAvailable checks if a name is available.
func (c *Client) NameAvailable(name string) (bool, error) {
	exists, err := deployment_repo.New().ExistsByName(name)
	if err != nil {
		return false, err
	}

	return !exists, nil
}

// createImagePath creates a complete container image path that can be pulled from.
func createImagePath(ownerID, name string) string {
	return fmt.Sprintf("%s/%s/%s", config.Config.Registry.URL, subsystemutils.GetPrefixedName(ownerID), name)
}

// createTransferCode generates a transfer code.
func createTransferCode() string {
	return utils.HashStringAlphanumeric(uuid.NewString())
}
