package deployments

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go-deploy/models/dto/v1/body"
	deploymentModels "go-deploy/models/sys/deployment"
	jobModels "go-deploy/models/sys/job"
	notificationModels "go-deploy/models/sys/notification"
	teamModels "go-deploy/models/sys/team"
	"go-deploy/models/versions"
	"go-deploy/pkg/config"
	sErrors "go-deploy/service/errors"
	utils2 "go-deploy/service/utils"
	"go-deploy/service/v1/deployments/gitlab_service"
	"go-deploy/service/v1/deployments/opts"
	"go-deploy/utils"
	"go-deploy/utils/subsystemutils"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"sort"
	"strings"
	"time"
)

// Get gets an existing deployment.
//
// It can be fetched in multiple ways including ID, name, transfer code, and Harbor webhook.
// It supports service.AuthInfo, and will restrict the result to ensure the user has access to the resource.
func (c *Client) Get(id string, opts ...opts.GetOpts) (*deploymentModels.Deployment, error) {
	o := utils2.GetFirstOrDefault(opts)

	dClient := deploymentModels.New()

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
		teamCheck, err = teamModels.New().WithUserID(c.V1.Auth().UserID).WithResourceID(id).ExistsAny()
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
// It supports service.AuthInfo, and will restrict the result to ensure the user has access to the resource.
func (c *Client) List(opts ...opts.ListOpts) ([]deploymentModels.Deployment, error) {
	o := utils2.GetFirstOrDefault(opts)

	dClient := deploymentModels.New()

	if o.Pagination != nil {
		dClient.WithPagination(o.Pagination.Page, o.Pagination.PageSize)
	}

	if o.GitHubWebhookID != nil {
		dClient.WithGitHubWebhookID(*o.GitHubWebhookID)
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

		teamClient := teamModels.New().WithUserID(effectiveUserID)
		if o.Pagination != nil {
			teamClient.WithPagination(o.Pagination.Page, o.Pagination.PageSize)
		}

		teams, err := teamClient.List()
		if err != nil {
			return nil, err
		}

		for _, team := range teams {
			for _, resource := range team.GetResourceMap() {
				if resource.Type != teamModels.ResourceTypeDeployment {
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

	params := &deploymentModels.CreateParams{}
	params.FromDTO(deploymentCreate, fallbackZone, fallbackImage, fallbackPort)

	d, err := deploymentModels.New().Create(id, ownerID, params)
	if err != nil {
		if errors.Is(err, deploymentModels.NonUniqueFieldErr) {
			return sErrors.NonUniqueFieldErr
		}

		return makeError(err)
	}

	if d == nil {
		return makeError(fmt.Errorf("deployment already exists for another user"))
	}

	if d.Type == deploymentModels.TypeCustom {
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

	createPlaceHolderInstead := false
	if params.GitHub != nil {
		err = c.GitHub().WithRepositoryID(deploymentCreate.GitHub.RepositoryID).WithToken(deploymentCreate.GitHub.Token).Create(id, params)
		if err != nil {
			errString := err.Error()
			if strings.Contains(errString, "/hooks: 404 Not Found") {
				utils.PrettyPrintError(makeError(fmt.Errorf("webhook api not found. assuming github is not supported, inserting placeholder instead")))
				createPlaceHolderInstead = true
			} else if strings.Contains(errString, "401 Bad credentials") {
				utils.PrettyPrintError(makeError(fmt.Errorf("bad credentials. assuming github credentials expired or were revoked, inserting placeholder instead")))
				createPlaceHolderInstead = true
			} else {
				return makeError(err)
			}
		}
	} else {
		createPlaceHolderInstead = true
	}

	if createPlaceHolderInstead {
		err = c.GitHub().CreatePlaceholder(id)
		if err != nil {
			return makeError(err)
		}
	}

	// fetch again to make sure all the fields are populated
	d, err = c.Refresh(id)
	if err != nil {
		return makeError(err)
	}

	if d.Subsystems.GitHub.Created() && params.GitHub != nil {
		gc := c.GitHub().WithRepositoryID(params.GitHub.RepositoryID).WithToken(params.GitHub.Token)
		repo, err := gc.GetRepository()
		if err != nil {
			return makeError(err)
		}

		if repo == nil {
			log.Println("github repository not found. assuming it was deleted")
			return nil
		}

		if repo.DefaultBranch == "" {
			log.Println("github repository has no default branch. assuming it was deleted")
			return nil
		}

		if repo.CloneURL == "" {
			log.Println("github repository has no clone url. assuming it was deleted")
			return nil
		}

		err = c.build([]string{id}, &deploymentModels.BuildParams{
			Name:      repo.Name,
			Tag:       "latest",
			Branch:    repo.DefaultBranch,
			ImportURL: repo.CloneURL,
		})
		if err != nil {
			return makeError(err)
		}
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

	params := &deploymentModels.UpdateParams{}
	params.FromDTO(dtoUpdate, d.Type)

	if params.Name != nil && d.Type == deploymentModels.TypeCustom {
		image := createImagePath(d.OwnerID, *params.Name)
		params.Image = &image
	}

	// Don't update the custom domain secret if the update contains the same domain
	if params.CustomDomain != nil && mainApp.CustomDomain != nil && *params.CustomDomain == mainApp.CustomDomain.Domain {
		params.CustomDomain = nil
	}

	err = deploymentModels.New().UpdateWithParams(id, params)
	if err != nil {
		if errors.Is(err, deploymentModels.NonUniqueFieldErr) {
			return sErrors.NonUniqueFieldErr
		}

		return makeError(err)
	}

	d, err = c.Refresh(id)
	if err != nil {
		return makeError(err)
	}

	if d.Type == deploymentModels.TypeCustom {
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
		err := c.V1.Jobs().Create(jobID, effectiveUserID, jobModels.TypeUpdateDeploymentOwner, versions.V1, map[string]interface{}{
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
	err = deploymentModels.New().UpdateWithParams(id, &deploymentModels.UpdateParams{
		TransferUserID: &params.NewOwnerID,
		TransferCode:   &code,
	})
	if err != nil {
		return nil, makeError(err)
	}

	_, err = c.V1.Notifications().Create(uuid.NewString(), params.NewOwnerID, &notificationModels.CreateParams{
		Type: notificationModels.TypeDeploymentTransfer,
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
	if d.Type == deploymentModels.TypeCustom {
		image := createImagePath(params.NewOwnerID, d.Name)
		newImage = &image
	}

	emptyString := ""

	err = deploymentModels.New().UpdateWithParams(id, &deploymentModels.UpdateParams{
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

	nmc := notificationModels.New().WithUserID(params.NewOwnerID).FilterContent("id", id).WithType(notificationModels.TypeDeploymentTransfer)
	err = nmc.MarkReadAndCompleted()
	if err != nil {
		return makeError(err)
	}

	log.Println("deployment", id, "owner updated from", params.OldOwnerID, "to", params.NewOwnerID)
	return nil
}

// ClearUpdateOwner clears the owner update process.
//
// This is intended to be used when the owner update process is cancelled.
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
	err = deploymentModels.New().UpdateWithParams(id, &deploymentModels.UpdateParams{
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

	nmc := notificationModels.New().FilterContent("id", id)
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

	err = c.GitHub().Delete(id)
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

// Restart restarts an deployment.
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

	c.AddLogs(id, deploymentModels.Log{
		Source: deploymentModels.LogSourceDeployment,
		Prefix: "[deployment]",
		// Since this is sent as a string, and not a JSON object, we need to prepend the createdAt
		Line:      fmt.Sprintf("%s %s", time.Now().Format(time.RFC3339), "Restart requested"),
		CreatedAt: time.Now(),
	})

	err = deploymentModels.New().SetWithBsonByID(id, bson.D{{"restartedAt", time.Now()}})
	if err != nil {
		return makeError(err)
	}

	err = c.StartActivity(id, deploymentModels.ActivityRestarting)
	if err != nil {
		return makeError(err)
	}

	go func() {
		// the restart is best-effort, so we mimic a reasonable delay
		time.Sleep(5 * time.Second)

		err = deploymentModels.New().RemoveActivity(id, deploymentModels.ActivityRestarting)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to remove activity %s from deployment %s. details: %w", deploymentModels.ActivityRestarting, id, err))
		}
	}()

	err = c.K8s().Restart(id)
	if err != nil {
		return makeError(err)
	}

	return nil
}

// Build builds an deployment.
//
// It can build by either a list of IDs or a single ID.
// Use WithID or WithIDs to set the ID(s) (prioritizes ID over IDs).
//
// It will filter out all the deployments that are not ready to build.
// Which means, all the deployments for supplied IDs might not be built.
func (c *Client) Build(ids []string, buildParams *body.DeploymentBuild) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to build deployment. details: %w", err)
	}

	params := &deploymentModels.BuildParams{}
	params.FromDTO(buildParams)

	for _, id := range ids {
		err := deploymentModels.New().AddLogs(id, deploymentModels.Log{
			Source: deploymentModels.LogSourceDeployment,
			Prefix: "[deployment]",
			// Since this is sent as a string, and not a JSON object, we need to prepend the createdAt
			Line:      fmt.Sprintf("%s %s", time.Now().Format(time.RFC3339), "Build requested"),
			CreatedAt: time.Now(),
		})
		if err != nil {
			return makeError(err)
		}
	}

	err := c.build(ids, params)
	if err != nil {
		return makeError(err)
	}

	return nil
}

// AddLogs adds logs to the deployment.
//
// It is purely done in best-effort
func (c *Client) AddLogs(id string, logs ...deploymentModels.Log) {
	// logs are added best-effort, so we don't return an error here
	go func() {
		err := deploymentModels.New().AddLogs(id, logs...)
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

	err := deploymentModels.New().AddActivity(id, activity)
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
	case deploymentModels.ActivityBeingCreated:
		return !d.BeingDeleted(), "Resource is being deleted"
	case deploymentModels.ActivityBeingDeleted:
		return true, ""
	case deploymentModels.ActivityUpdating:
		return !d.BeingDeleted() && !d.BeingCreated(), "Resource is being deleted or created"
	case deploymentModels.ActivityRestarting:
		return !d.BeingDeleted(), "Resource is being deleted"
	case deploymentModels.ActivityBuilding:
		return !d.BeingDeleted(), "Resource is being deleted"
	case deploymentModels.ActivityRepairing:
		return d.Ready(), "Resource is not ready"
	}

	return false, fmt.Sprintf("Unknown activity %s", activity)
}

// GetUsage gets the usage of the user.
func (c *Client) GetUsage(userID string) (*deploymentModels.Usage, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get usage. details: %w", err)
	}

	count, err := deploymentModels.New().WithOwner(userID).CountReplicas()
	if err != nil {
		return nil, makeError(err)
	}

	return &deploymentModels.Usage{
		Count: count,
	}, nil
}

// NameAvailable checks if a name is available.
func (c *Client) NameAvailable(name string) (bool, error) {
	exists, err := deploymentModels.New().ExistsByName(name)
	if err != nil {
		return false, err
	}

	return !exists, nil
}

// ValidGitHubToken validates a GitHub token.
//
// It returns a boolean indicating if the token is valid, a string indicating the reason if it is not, and an error if any.
func (c *Client) ValidGitHubToken(token string) (bool, string, error) {
	return c.GitHub().WithToken(token).Validate()
}

// GetGitHubAccessTokenByCode gets a GitHub access token by a code.
func (c *Client) GetGitHubAccessTokenByCode(code string) (string, error) {
	code, err := c.GitHub().GetAccessTokenByCode(code)
	if err != nil {
		utils.PrettyPrintError(fmt.Errorf("failed to get github access token. details: %w", err))
		return "", err
	}

	return code, nil
}

// GetGitHubRepositories gets GitHub repositories for a token.
//
// The token should be validated before calling this function.
// If the token is expired, an error will be returned.
func (c *Client) GetGitHubRepositories(token string) ([]deploymentModels.GitHubRepository, error) {
	return c.GitHub().WithToken(token).GetRepositories()
}

// ValidGitHubRepository validates a GitHub repository.
//
// The token should be validated before calling this function.
// If the token is expired, an error will be returned.
func (c *Client) ValidGitHubRepository(token string, repositoryID int64) (bool, string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get github repository. details: %w", err)
	}

	gc := c.GitHub().WithToken(token).WithRepositoryID(repositoryID)

	repo, err := c.GitHub().WithRepositoryID(repositoryID).WithToken(token).GetRepository()
	if err != nil {
		return false, "", makeError(err)
	}

	if repo == nil {
		return false, "Repository not found", nil
	}

	webhooks, err := gc.GetWebhooks(repo)
	if err != nil {
		return false, "", makeError(err)
	}

	webhooksWithPushEvent := make([]*deploymentModels.GitHubWebhook, 0)
	for _, webhook := range webhooks {
		for _, event := range webhook.Events {
			if event == "push" {
				webhooksWithPushEvent = append(webhooksWithPushEvent, &webhook)
			}
		}
	}

	if len(webhooksWithPushEvent) >= 20 {
		return false, "Too many webhooks with push event", nil
	}

	return true, "", nil
}

// build builds a deployment.
//
// It is a helper function that does not do any checks.
func (c *Client) build(ids []string, params *deploymentModels.BuildParams) error {
	var filtered []string
	for _, id := range ids {
		err := c.StartActivity(id, deploymentModels.ActivityBuilding)
		if err != nil {
			var failedToStartActivityErr sErrors.FailedToStartActivityError
			if errors.As(err, &failedToStartActivityErr) {
				log.Println("could not start building activity for deployment", id, ". reason:", failedToStartActivityErr.Error())
				continue
			}

			if errors.Is(err, sErrors.DeploymentNotFoundErr) {
				log.Println("deployment", id, "not found when starting activity", deploymentModels.ActivityBuilding, ". assuming it was deleted")
				continue
			}

			return err
		}

		filtered = append(filtered, id)
	}
	defer func() {
		for _, id := range filtered {
			err := deploymentModels.New().RemoveActivity(id, deploymentModels.ActivityBuilding)
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("failed to remove activity %s for deployment %s. details: %w", deploymentModels.ActivityBuilding, id, err))
			}
		}
	}()

	if len(filtered) == 0 {
		return nil
	}

	err := gitlab_service.CreateBuild(filtered, params)
	if err != nil {
		// we treat building as a non-critical activity, so we don't return an error here
		utils.PrettyPrintError(fmt.Errorf("failed to build image for %d deployments. details: %w", len(filtered), err))
	}

	return nil
}

// createImagePath creates a complete container image path that can be pulled from.
func createImagePath(ownerID, name string) string {
	return fmt.Sprintf("%s/%s/%s", config.Config.Registry.URL, subsystemutils.GetPrefixedName(ownerID), name)
}

// createTransferCode generates a transfer code.
func createTransferCode() string {
	return utils.HashStringAlphanumeric(uuid.NewString())
}
