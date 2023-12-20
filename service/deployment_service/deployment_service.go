package deployment_service

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go-deploy/models/dto/body"
	deploymentModel "go-deploy/models/sys/deployment"
	jobModel "go-deploy/models/sys/job"
	notificationModel "go-deploy/models/sys/notification"
	teamModels "go-deploy/models/sys/team"
	"go-deploy/pkg/config"
	"go-deploy/service/deployment_service/client"
	"go-deploy/service/deployment_service/github_service"
	"go-deploy/service/deployment_service/gitlab_service"
	"go-deploy/service/deployment_service/harbor_service"
	"go-deploy/service/deployment_service/k8s_service"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/job_service"
	"go-deploy/service/notification_service"
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
func (c *Client) Get(id string, opts *client.GetOptions) (*deploymentModel.Deployment, error) {
	dClient := deploymentModel.New()

	if opts.TransferCode != "" {
		return dClient.GetByTransferCode(opts.TransferCode)
	}

	var effectiveUserID string
	if c.Auth != nil && !c.Auth.IsAdmin {
		effectiveUserID = c.Auth.UserID
	}

	var teamCheck bool
	if !opts.Shared {
		teamCheck = false
	} else if c.Auth == nil || c.Auth.IsAdmin {
		teamCheck = true
	} else {
		var err error
		teamCheck, err = teamModels.New().WithUserID(c.Auth.UserID).WithResourceID(id).ExistsAny()
		if err != nil {
			return nil, err
		}
	}

	if !teamCheck && effectiveUserID != "" {
		dClient.RestrictToOwner(effectiveUserID)
	}

	if opts.HarborWebhook != nil {
		return dClient.GetByName(opts.HarborWebhook.EventData.Repository.Name)
	}

	return c.Deployment(id, dClient)
}

// List lists existing deployments.
//
// It supports service.AuthInfo, and will restrict the result to ensure the user has access to the resource.
func (c *Client) List(opts *client.ListOptions) ([]deploymentModel.Deployment, error) {
	dClient := deploymentModel.New()

	if opts.Pagination != nil {
		dClient.WithPagination(opts.Pagination.Page, opts.Pagination.PageSize)
	}

	if opts.GitHubWebhookID != 0 {
		dClient.WithGitHubWebhookID(opts.GitHubWebhookID)
	}

	var effectiveUserID string
	if opts.UserID != "" {
		// Specific user's deployments are requested
		if c.Auth == nil || c.Auth.UserID == opts.UserID || c.Auth.IsAdmin {
			effectiveUserID = opts.UserID
		} else {
			// User cannot access the other user's resources
			effectiveUserID = c.Auth.UserID
		}
	} else {
		// All deployments are requested
		if c.Auth != nil && !c.Auth.IsAdmin {
			effectiveUserID = c.Auth.UserID
		}
	}

	if effectiveUserID != "" {
		dClient.RestrictToOwner(effectiveUserID)
	}

	resources, err := c.Deployments(dClient)
	if err != nil {
		return nil, err
	}

	// Can only view shared if we are listing resources for a specific user
	if opts.Shared && effectiveUserID != "" {
		skipIDs := make([]string, len(resources))
		for i, resource := range resources {
			skipIDs[i] = resource.ID
		}

		teamClient := teamModels.New().WithUserID(effectiveUserID)
		if opts.Pagination != nil {
			teamClient.WithPagination(opts.Pagination.Page, opts.Pagination.PageSize)
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
		if opts.Pagination != nil {
			resources = utils.GetPage(resources, opts.Pagination.PageSize, opts.Pagination.Page)
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

	params := &deploymentModel.CreateParams{}
	params.FromDTO(deploymentCreate, fallbackZone, fallbackImage, fallbackPort)

	d, err := deploymentModel.New().Create(id, ownerID, params)
	if err != nil {
		if errors.Is(err, deploymentModel.NonUniqueFieldErr) {
			return sErrors.NonUniqueFieldErr
		}

		return makeError(err)
	}

	if d == nil {
		return makeError(fmt.Errorf("deployment already exists for another user"))
	}

	if d.Type == deploymentModel.TypeCustom {
		err = harbor_service.New(c.Cache).Create(id, params)
		if err != nil {
			return makeError(err)
		}
	} else {
		err = harbor_service.New(c.Cache).CreatePlaceholder(id)
		if err != nil {
			return makeError(err)
		}
	}

	d, err = c.Refresh(id)
	if err != nil {
		return makeError(err)
	}

	err = k8s_service.New(c.Cache).Create(id, params)
	if err != nil {
		return makeError(err)
	}

	createPlaceHolderInstead := false
	if params.GitHub != nil {
		err = github_service.New(c.Cache).WithRepositoryID(deploymentCreate.GitHub.RepositoryID).WithToken(deploymentCreate.GitHub.Token).Create(id, params)
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
		err = github_service.New(c.Cache).CreatePlaceholder(id)
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
		gc := github_service.New(c.Cache).WithRepositoryID(params.GitHub.RepositoryID).WithToken(params.GitHub.Token)
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

		err = c.build([]string{id}, &deploymentModel.BuildParams{
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

	params := &deploymentModel.UpdateParams{}
	params.FromDTO(dtoUpdate, d.Type)

	if params.Name != nil && d.Type == deploymentModel.TypeCustom {
		image := createImagePath(d.OwnerID, *params.Name)
		params.Image = &image
	}

	err = deploymentModel.New().UpdateWithParamsByID(id, params)
	if err != nil {
		if errors.Is(err, deploymentModel.NonUniqueFieldErr) {
			return sErrors.NonUniqueFieldErr
		}

		return makeError(err)
	}

	if d.Type == deploymentModel.TypeCustom {
		err = harbor_service.New(c.Cache).Update(id, params)
		if err != nil {
			return makeError(err)
		}
	}

	err = k8s_service.New(c.Cache).Update(id, params)
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

	if c.Auth == nil || c.Auth.IsAdmin {
		doTransfer = true
	} else if c.Auth.UserID == params.NewOwnerID {
		if params.TransferCode == nil || d.Transfer == nil || d.Transfer.Code != *params.TransferCode {
			return nil, sErrors.InvalidTransferCodeErr
		}

		doTransfer = true
	}

	var effectiveUserID string
	if c.Auth == nil {
		effectiveUserID = "system"
	} else {
		effectiveUserID = c.Auth.UserID
	}

	if doTransfer {
		jobID := uuid.New().String()
		err := job_service.Create(jobID, effectiveUserID, jobModel.TypeUpdateDeploymentOwner, map[string]interface{}{
			"id":     id,
			"params": *params,
		})

		if err != nil {
			return nil, makeError(err)
		}

		return &jobID, nil
	}

	/// create a transfer notification
	code := createTransferCode()
	err = deploymentModel.New().UpdateWithParamsByID(id, &deploymentModel.UpdateParams{
		TransferUserID: &params.NewOwnerID,
		TransferCode:   &code,
	})
	if err != nil {
		return nil, makeError(err)
	}

	err = notification_service.CreateNotification(uuid.NewString(), params.NewOwnerID, &notificationModel.CreateParams{
		Type: notificationModel.TypeDeploymentTransfer,
		Content: map[string]interface{}{
			"id":     d.ID,
			"name":   d.Name,
			"userId": params.OldOwnerID,
			"email":  c.Auth.GetEmail(),
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
	if d.Type == deploymentModel.TypeCustom {
		image := createImagePath(params.NewOwnerID, d.Name)
		newImage = &image
	}

	emptyString := ""

	err = deploymentModel.New().UpdateWithParamsByID(id, &deploymentModel.UpdateParams{
		OwnerID:        &params.NewOwnerID,
		Image:          newImage,
		TransferCode:   &emptyString,
		TransferUserID: &emptyString,
	})
	if err != nil {
		return makeError(err)
	}

	err = harbor_service.New(c.Cache).EnsureOwner(id, params.OldOwnerID)
	if err != nil {
		return makeError(err)
	}

	err = k8s_service.New(c.Cache).EnsureOwner(id, params.OldOwnerID)
	if err != nil {
		return makeError(err)
	}

	log.Println("deployment", id, "owner updated to", params.NewOwnerID)
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
	err = deploymentModel.New().UpdateWithParamsByID(id, &deploymentModel.UpdateParams{
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

	err := harbor_service.New(c.Cache).Delete(id)
	if err != nil {
		return makeError(err)
	}

	err = k8s_service.New(c.Cache).Delete(id)
	if err != nil {
		return makeError(err)
	}

	err = github_service.New(c.Cache).Delete(id)
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

	err = k8s_service.New(c.Cache).Repair(id)
	if err != nil {
		if errors.Is(err, sErrors.IngressHostInUseErr) {
			// The user should fix this error, so we don't return an error here
			utils.PrettyPrintError(err)
		} else {
			return makeError(err)
		}
	}

	if !d.Subsystems.Harbor.Placeholder {
		err = harbor_service.New(c.Cache).Repair(id)
		if err != nil {
			return makeError(err)
		}
	}

	log.Println("repaired deployment", id)
	return nil
}

// Restart restarts an existing deployment.
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

	c.AddLogs(id, deploymentModel.Log{
		Source:    deploymentModel.LogSourceDeployment,
		Prefix:    "[deployment]",
		Line:      "Restart requested",
		CreatedAt: time.Now(),
	})

	err = deploymentModel.New().SetWithBsonByID(id, bson.D{{"restartedAt", time.Now()}})
	if err != nil {
		return makeError(err)
	}

	err = c.StartActivity(id, deploymentModel.ActivityRestarting)
	if err != nil {
		return makeError(err)
	}

	go func() {
		// the restart is best-effort, so we mimic a reasonable delay
		time.Sleep(5 * time.Second)

		err = deploymentModel.New().RemoveActivity(id, deploymentModel.ActivityRestarting)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to remove activity %s from deployment %s. details: %w", deploymentModel.ActivityRestarting, id, err))
		}
	}()

	err = k8s_service.New(c.Cache).Restart(id)
	if err != nil {
		return makeError(err)
	}

	return nil
}

// Build builds an existing deployment.
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

	params := &deploymentModel.BuildParams{}
	params.FromDTO(buildParams)

	for _, id := range ids {
		err := deploymentModel.New().AddLogs(id, deploymentModel.Log{
			Source:    deploymentModel.LogSourceDeployment,
			Prefix:    "[deployment]",
			Line:      "Build requested",
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
func (c *Client) AddLogs(id string, logs ...deploymentModel.Log) {
	// logs are added best-effort, so we don't return an error here
	go func() {
		err := deploymentModel.New().AddLogs(id, logs...)
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
func (c *Client) CheckQuota(id string, opts *client.QuotaOptions) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check quota. details: %w", err)
	}

	if c.Auth == nil || c.Auth.IsAdmin {
		return nil
	}

	usage, err := c.GetUsage(c.Auth.UserID)
	if err != nil {
		return makeError(err)
	}

	quota := c.Auth.GetEffectiveRole().Quotas

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

	err := deploymentModel.New().AddActivity(id, activity)
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
	case deploymentModel.ActivityBeingCreated:
		return !d.BeingDeleted(), "Resource is being deleted"
	case deploymentModel.ActivityBeingDeleted:
		return true, ""
	case deploymentModel.ActivityUpdating:
		return !d.BeingDeleted() && !d.BeingCreated(), "Resource is being deleted or created"
	case deploymentModel.ActivityRestarting:
		return !d.BeingDeleted(), "Resource is being deleted"
	case deploymentModel.ActivityBuilding:
		return !d.BeingDeleted(), "Resource is being deleted"
	case deploymentModel.ActivityRepairing:
		return d.Ready(), "Resource is not ready"
	}

	return false, fmt.Sprintf("Unknown activity %s", activity)
}

// GetUsage gets the usage of the user.
func (c *Client) GetUsage(userID string) (*deploymentModel.Usage, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get usage. details: %w", err)
	}

	count, err := deploymentModel.New().RestrictToOwner(userID).CountReplicas()
	if err != nil {
		return nil, makeError(err)
	}

	return &deploymentModel.Usage{
		Count: count,
	}, nil
}

// SavePing saves a ping result.
func (c *Client) SavePing(id string, pingResult int) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update deployment with ping result. details: %w", err)
	}

	deployment, err := c.Deployment(id, nil)
	if err != nil {
		return makeError(err)
	}

	if deployment == nil {
		log.Println("deployment", id, "not found when updating ping result. assuming it was deleted")
		return nil
	}

	err = deploymentModel.New().SavePing(id, pingResult)
	if err != nil {
		return makeError(err)
	}

	return nil
}

// build builds a deployment.
//
// It is a helper function that does not do any checks.
func (c *Client) build(ids []string, params *deploymentModel.BuildParams) error {
	var filtered []string
	for _, id := range ids {
		err := c.StartActivity(id, deploymentModel.ActivityBuilding)
		if err != nil {
			var failedToStartActivityErr sErrors.FailedToStartActivityError
			if errors.As(err, &failedToStartActivityErr) {
				log.Println("could not start building activity for deployment", id, ". reason:", failedToStartActivityErr.Error())
				continue
			}

			if errors.Is(err, sErrors.DeploymentNotFoundErr) {
				log.Println("deployment", id, "not found when starting activity", deploymentModel.ActivityBuilding, ". assuming it was deleted")
				continue
			}

			return err
		}

		filtered = append(filtered, id)
	}
	defer func() {
		for _, id := range filtered {
			err := deploymentModel.New().RemoveActivity(id, deploymentModel.ActivityBuilding)
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("failed to remove activity %s for deployment %s. details: %w", deploymentModel.ActivityBuilding, id, err))
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

// ValidGitHubToken validates a GitHub token.
//
// It returns a boolean indicating if the token is valid, a string indicating the reason if it is not, and an error if any.
func ValidGitHubToken(token string) (bool, string, error) {
	return github_service.New(nil).WithToken(token).Validate()
}

// GetGitHubAccessTokenByCode gets a GitHub access token by a code.
func GetGitHubAccessTokenByCode(code string) (string, error) {
	code, err := github_service.GetAccessTokenByCode(code)
	if err != nil {
		utils.PrettyPrintError(fmt.Errorf("failed to get github access token. details: %w", err))
		return "", err
	}

	return code, nil
}

// NameAvailable checks if a name is available.
func NameAvailable(name string) (bool, error) {
	exists, err := deploymentModel.New().ExistsByName(name)
	if err != nil {
		return false, err
	}

	return !exists, nil
}

// GetGitHubRepositories gets GitHub repositories for a token.
//
// The token should be validated before calling this function.
// If the token is expired, an error will be returned.
func GetGitHubRepositories(token string) ([]deploymentModel.GitHubRepository, error) {
	return github_service.New(nil).WithToken(token).GetRepositories()
}

// ValidGitHubRepository validates a GitHub repository.
//
// The token should be validated before calling this function.
// If the token is expired, an error will be returned.
func ValidGitHubRepository(token string, repositoryID int64) (bool, string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get github repository. details: %w", err)
	}

	gc := github_service.New(nil).WithToken(token).WithRepositoryID(repositoryID)

	repo, err := github_service.New(nil).WithRepositoryID(repositoryID).WithToken(token).GetRepository()
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

	webhooksWithPushEvent := make([]*deploymentModel.GitHubWebhook, 0)
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

// createImagePath creates a complete container image path that can be pulled from.
func createImagePath(ownerID, name string) string {
	return fmt.Sprintf("%s/%s/%s", config.Config.Registry.URL, subsystemutils.GetPrefixedName(ownerID), name)
}

// createTransferCode generates a transfer code.
func createTransferCode() string {
	return utils.HashString(uuid.NewString())
}
