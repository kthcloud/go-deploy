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
	"go-deploy/service"
	"go-deploy/service/deployment_service/base"
	"go-deploy/service/deployment_service/client"
	dErrors "go-deploy/service/deployment_service/errors"
	"go-deploy/service/deployment_service/github_service"
	"go-deploy/service/deployment_service/gitlab_service"
	"go-deploy/service/deployment_service/harbor_service"
	"go-deploy/service/deployment_service/k8s_service"
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

func (c *Client) Create(deploymentID, ownerID string, deploymentCreate *body.DeploymentCreate) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create deployment. details: %w", err)
	}

	// temporary hard-coded fallback
	fallbackZone := "se-flem"
	fallbackImage := createImagePath(ownerID, deploymentCreate.Name)
	fallbackPort := config.Config.Deployment.Port

	params := &deploymentModel.CreateParams{}
	params.FromDTO(deploymentCreate, fallbackZone, fallbackImage, fallbackPort)

	deployment, err := deploymentModel.New().Create(deploymentID, ownerID, params)
	if err != nil {
		if errors.Is(err, deploymentModel.NonUniqueFieldErr) {
			return dErrors.NonUniqueFieldErr
		}

		return makeError(err)
	}

	if deployment == nil {
		return makeError(fmt.Errorf("deployment already exists for another user"))
	}

	if deployment.Type == deploymentModel.TypeCustom {
		err = harbor_service.Create(deploymentID, params)
		if err != nil {
			return makeError(err)
		}
	} else {
		err = harbor_service.CreatePlaceholder(deploymentID)
		if err != nil {
			return makeError(err)
		}
	}

	err = k8s_service.Create(deploymentID, params)
	if err != nil {
		if errors.Is(err, base.CustomDomainInUseErr) {
			log.Println("custom domain in use when creating deployment", params.Name, ". removing it from the deployment and create params")
			err = deploymentModel.New().RemoveCustomDomain(deploymentID)
			if err != nil {
				return makeError(err)
			}
			params.CustomDomain = nil
		} else {
			return makeError(err)
		}
	}

	createPlaceHolderInstead := false
	if params.GitHub != nil {
		err = github_service.Create(deploymentID, params)
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
		err = github_service.CreatePlaceholder(deploymentID)
		if err != nil {
			return makeError(err)
		}
	}

	// fetch again to make sure all the fields are populated
	err = c.Fetch()
	if err != nil {
		return makeError(err)
	}

	if deployment.Subsystems.GitHub.Created() && params.GitHub != nil {
		repo, err := github_service.GetRepository(params.GitHub.Token, params.GitHub.RepositoryID)
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

		err = c.build(&deploymentModel.BuildParams{
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

func (c *Client) Update(dtoUpdate *body.DeploymentUpdate) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update deployment. details: %w", err)
	}

	if c.Deployment() == nil {
		return dErrors.DeploymentNotFoundErr
	}

	params := &deploymentModel.UpdateParams{}
	params.FromDTO(dtoUpdate, c.Deployment().Type)

	if params.Name != nil && c.Deployment().Type == deploymentModel.TypeCustom {
		image := createImagePath(c.Deployment().OwnerID, *params.Name)
		params.Image = &image
	}

	err := deploymentModel.New().UpdateWithParamsByID(c.ID(), params)
	if err != nil {
		if errors.Is(err, deploymentModel.NonUniqueFieldErr) {
			return dErrors.NonUniqueFieldErr
		}

		return makeError(err)
	}

	if c.Deployment().Type == deploymentModel.TypeCustom {
		err = harbor_service.Update(c.ID(), params)
		if err != nil {
			return makeError(err)
		}
	}

	err = k8s_service.Update(c.ID(), params)
	if err != nil {
		if errors.Is(err, base.CustomDomainInUseErr) {
			log.Println("custom domain in use when updating deployment", c.Deployment().Name, ". removing it from the update params")
			dtoUpdate.CustomDomain = nil
		} else {
			return makeError(err)
		}
	}

	return nil
}

func (c *Client) UpdateOwnerSetup(params *body.DeploymentUpdateOwner) (*string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update deployment owner. details: %w", err)
	}

	if c.Deployment() == nil {
		return nil, dErrors.DeploymentNotFoundErr
	}

	if c.Deployment().OwnerID == params.NewOwnerID {
		return nil, nil
	}

	doTransfer := false

	if c.Auth == nil || c.Auth.IsAdmin {
		doTransfer = true
	} else if c.Auth.UserID == params.NewOwnerID {
		if params.TransferCode == nil || c.Deployment().Transfer == nil || c.Deployment().Transfer.Code != *params.TransferCode {
			return nil, dErrors.InvalidTransferCodeErr
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
			"id":     c.ID(),
			"params": *params,
		})

		if err != nil {
			return nil, makeError(err)
		}

		return &jobID, nil
	}

	/// create a transfer notification
	code := createTransferCode()
	err := deploymentModel.New().UpdateWithParamsByID(c.ID(), &deploymentModel.UpdateParams{
		TransferUserID: &params.NewOwnerID,
		TransferCode:   &code,
	})
	if err != nil {
		return nil, makeError(err)
	}

	err = notification_service.CreateNotification(uuid.NewString(), params.NewOwnerID, &notificationModel.CreateParams{
		Type: notificationModel.TypeDeploymentTransfer,
		Content: map[string]interface{}{
			"id":     c.Deployment().ID,
			"name":   c.Deployment().Name,
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

func (c *Client) ClearUpdateOwner() error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to clear deployment owner update. details: %w", err)
	}

	if c.Deployment() == nil {
		return dErrors.DeploymentNotFoundErr
	}

	if c.Deployment().Transfer == nil {
		return nil
	}

	emptyString := ""
	err := deploymentModel.New().UpdateWithParamsByID(c.ID(), &deploymentModel.UpdateParams{
		TransferUserID: &emptyString,
		TransferCode:   &emptyString,
	})
	if err != nil {
		return makeError(err)
	}

	// TODO: delete notification?

	return nil
}

func (c *Client) UpdateOwner(params *body.DeploymentUpdateOwner) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update deployment owner. details: %w", err)
	}

	if c.Deployment() == nil {
		return dErrors.DeploymentNotFoundErr
	}

	var newImage *string
	if c.Deployment().Type == deploymentModel.TypeCustom {
		image := createImagePath(params.NewOwnerID, c.Deployment().Name)
		newImage = &image
	}

	emptyString := ""

	err := deploymentModel.New().UpdateWithParamsByID(c.ID(), &deploymentModel.UpdateParams{
		OwnerID:        &params.NewOwnerID,
		Image:          newImage,
		TransferCode:   &emptyString,
		TransferUserID: &emptyString,
	})
	if err != nil {
		return makeError(err)
	}

	err = harbor_service.EnsureOwner(c.ID(), params.OldOwnerID)
	if err != nil {
		return makeError(err)
	}

	err = k8s_service.EnsureOwner(c.ID(), params.OldOwnerID)
	if err != nil {
		return makeError(err)
	}

	log.Println("deployment", c.ID(), "owner updated to", params.NewOwnerID)
	return nil
}

func (c *Client) Delete() error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete deployment. details: %w", err)
	}

	if !c.HasID() {
		return dErrors.DeploymentNotFoundErr
	}

	err := harbor_service.Delete(c.ID())
	if err != nil {
		return makeError(err)
	}

	err = k8s_service.Delete(c.ID())
	if err != nil {
		return makeError(err)
	}

	err = github_service.Delete(c.ID())
	if err != nil {
		return makeError(err)
	}

	return nil
}

func (c *Client) Repair() error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to repair deployment %s. details: %w", c.ID(), err)
	}

	if c.Deployment() == nil {
		return dErrors.DeploymentNotFoundErr
	}

	if !c.Deployment().Ready() {
		log.Println("deployment", c.ID(), "not ready when repairing.")
		return nil
	}

	err := k8s_service.Repair(c.ID())
	if err != nil {
		if errors.Is(err, base.CustomDomainInUseErr) {
			log.Println("custom domain in use when repairing deployment", c.ID(), ". removing it from the deployment")
			err = deploymentModel.New().RemoveCustomDomain(c.ID())
			if err != nil {
				return makeError(err)
			}
		} else {
			return makeError(err)
		}
	}

	if !c.Deployment().Subsystems.Harbor.Placeholder {
		err = harbor_service.Repair(c.ID())
		if err != nil {
			return makeError(err)
		}
	}

	log.Println("repaired deployment", c.ID())
	return nil
}

func (c *Client) Restart() error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to restart deployment. details: %w", err)
	}

	if c.Deployment() == nil {

	}

	c.AddLogs(deploymentModel.Log{
		Source:    deploymentModel.LogSourceDeployment,
		Prefix:    "[deployment]",
		Line:      "Restart requested",
		CreatedAt: time.Now(),
	})

	err := deploymentModel.New().SetWithBsonByID(c.ID(), bson.D{{"restartedAt", time.Now()}})
	if err != nil {
		return makeError(err)
	}

	started, reason, err := c.StartActivity(deploymentModel.ActivityRestarting)
	if err != nil {
		return makeError(err)
	}

	go func() {
		// the restart is best-effort, so we mimic a reasonable delay
		time.Sleep(5 * time.Second)

		err = deploymentModel.New().RemoveActivity(c.ID(), deploymentModel.ActivityRestarting)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to remove activity %s from deployment %s. details: %w", deploymentModel.ActivityRestarting, c.ID(), err))
		}
	}()

	if !started {
		return fmt.Errorf("failed to restart deployment %s. details: %s", c.ID(), reason)
	}

	err = k8s_service.Restart(c.ID())
	if err != nil {
		return makeError(err)
	}

	return nil
}

func (c *Client) Build(buildParams *body.DeploymentBuild) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to build deployment. details: %w", err)
	}

	params := &deploymentModel.BuildParams{}
	params.FromDTO(buildParams)

	var ids []string
	if c.HasID() {
		ids = []string{c.ID()}
	} else {
		ids = c.IDs
	}

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

	err := c.build(params)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func (c *Client) AddLogs(logs ...deploymentModel.Log) {
	// logs are added best-effort, so we don't return an error here
	go func() {
		err := deploymentModel.New().AddLogs(c.ID(), logs...)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to add logs to deployment %s. details: %w", c.ID(), err))
		}
	}()
}

func (c *Client) DoCommand(command string) {
	go func() {
		switch command {
		case "restart":
			err := c.Restart()
			if err != nil {
				utils.PrettyPrintError(err)
			}
		}
	}()
}

func (c *Client) Get(opts *client.GetOptions) (*deploymentModel.Deployment, error) {
	dClient := deploymentModel.New()

	if opts.TransferCode != "" {
		return dClient.GetByTransferCode(opts.TransferCode)
	}

	var effectiveUserID string
	if c.UserID != "" {
		if c.Auth == nil || c.Auth.UserID == c.UserID || c.Auth.IsAdmin {
			effectiveUserID = c.UserID
		} else {
			effectiveUserID = c.Auth.UserID
		}
	} else {
		if c.Auth != nil && !c.Auth.IsAdmin {
			effectiveUserID = c.Auth.UserID
		}
	}

	var teamCheck bool
	if !opts.Shared {
		teamCheck = false
	} else if c.Auth == nil || c.Auth.IsAdmin {
		teamCheck = true
	} else {
		var err error
		teamCheck, err = teamModels.New().AddUserID(c.Auth.UserID).AddResourceID(c.ID()).ExistsAny()
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

	return c.Deployment(), nil
}

func (c *Client) List(opts *client.ListOptions) ([]deploymentModel.Deployment, error) {
	dClient := deploymentModel.New()

	if opts.Pagination != nil {
		dClient.WithPagination(opts.Pagination.Page, opts.Pagination.PageSize)
	}

	if opts.GitHubWebhookID != 0 {
		dClient.WithGitHubWebhookID(opts.GitHubWebhookID)
	}

	var effectiveUserID string
	if c.UserID != "" {
		// specific user's deployments are requested
		if c.Auth == nil || c.Auth.UserID == c.UserID || c.Auth.IsAdmin {
			effectiveUserID = c.UserID
		} else {
			effectiveUserID = c.Auth.UserID
		}
	} else {
		// all deployments are requested
		if c.Auth != nil && !c.Auth.IsAdmin {
			effectiveUserID = c.Auth.UserID
		}
	}

	if effectiveUserID != "" {
		dClient.RestrictToOwner(effectiveUserID)
	}

	resources, err := dClient.List()
	if err != nil {
		return nil, err
	}

	if opts.Shared && effectiveUserID != "" {
		skipIDs := make([]string, len(resources))
		for i, resource := range resources {
			skipIDs[i] = resource.ID
		}

		teamClient := teamModels.New().AddUserID(effectiveUserID)
		if opts.Pagination != nil {
			teamClient.AddPagination(opts.Pagination.Page, opts.Pagination.PageSize)
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

				// skip existing non-shared resources
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

				deployment, err := deploymentModel.New().GetByID(resource.ID)
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

		// since we fetched from two collections, we need to do pagination manually
		if opts.Pagination != nil {
			resources = utils.GetPage(resources, opts.Pagination.PageSize, opts.Pagination.Page)
		}

	} else {
		// sort by createdAt
		sort.Slice(resources, func(i, j int) bool {
			return resources[i].CreatedAt.After(resources[j].CreatedAt)
		})
	}

	return resources, nil
}

func (c *Client) CheckQuota(opts *client.QuotaOptions) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check quota. details: %w", err)
	}

	if c.Auth != nil && c.Auth.IsAdmin {
		return nil
	}

	usage, err := c.GetUsage()
	if err != nil {
		return makeError(err)
	}

	if opts.Create != nil {
		add := 1
		if opts.Create.Replicas != nil {
			add = *opts.Create.Replicas
		}

		totalCount := usage.Count + add

		if totalCount > opts.Quota.Deployments {
			return service.NewQuotaExceededError(fmt.Sprintf("Deployment quota exceeded. Current: %d, Quota: %d", totalCount, opts.Quota.Deployments))
		}

		return nil
	} else if opts.Update != nil {
		add := 1
		if opts.Update.Replicas != nil {
			add = *opts.Update.Replicas
		}

		totalCount := usage.Count + add

		if totalCount > opts.Quota.Deployments {
			return service.NewQuotaExceededError(fmt.Sprintf("Deployment quota exceeded. Current: %d, Quota: %d", totalCount, opts.Quota.Deployments))
		}

		return nil
	} else {
		log.Println("quota options not set when checking quota for deployment", c.ID())
	}

	return nil
}

func (c *Client) StartActivity(activity string) (bool, string, error) {
	if !c.HasID() {
		return false, "Deployment not found", nil
	}

	canAdd, reason := c.CanAddActivity(activity)
	if !canAdd {
		return false, reason, nil
	}

	err := deploymentModel.New().AddActivity(c.ID(), activity)
	if err != nil {
		return false, "", err
	}

	return true, "", nil
}

func (c *Client) CanAddActivity(activity string) (bool, string) {
	d := c.Deployment()

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

func (c *Client) GetUsage() (*deploymentModel.Usage, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get usage. details: %w", err)
	}

	count, err := deploymentModel.New().RestrictToOwner(c.UserID).CountReplicas()
	if err != nil {
		return nil, makeError(err)
	}

	return &deploymentModel.Usage{
		Count: count,
	}, nil
}

func ValidGitHubToken(token string) (bool, string, error) {
	return github_service.ValidateToken(token)
}

func GetGitHubAccessTokenByCode(code string) (string, error) {
	code, err := github_service.GetAccessTokenByCode(code)
	if err != nil {
		utils.PrettyPrintError(fmt.Errorf("failed to get github access token. details: %w", err))
		return "", err
	}

	return code, nil
}

func NameAvailable(name string) (bool, error) {
	exists, err := deploymentModel.New().ExistsByName(name)
	if err != nil {
		return false, err
	}

	return !exists, nil
}

func GetGitHubRepositories(token string) ([]deploymentModel.GitHubRepository, error) {
	return github_service.GetRepositories(token)
}

func ValidGitHubRepository(token string, repositoryID int64) (bool, string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get github repository. details: %w", err)
	}

	repo, err := github_service.GetRepository(token, repositoryID)
	if err != nil {
		return false, "", makeError(err)
	}

	if repo == nil {
		return false, "Repository not found", nil
	}

	webhooks, err := github_service.GetWebhooks(token, repo.Owner, repo.Name)
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

func SavePing(id string, pingResult int) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update deployment with ping result. details: %w", err)
	}

	deployment, err := deploymentModel.New().GetByID(id)
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

func (c *Client) build(params *deploymentModel.BuildParams) error {
	var ids []string
	if c.HasID() {
		ids = []string{c.ID()}
	} else {
		ids = c.IDs
	}

	var filtered []string
	for _, id := range ids {
		started, reason, err := c.StartActivity(deploymentModel.ActivityBuilding)
		if err != nil {
			return err
		}

		if !started {
			utils.PrettyPrintError(fmt.Errorf("failed to build deployment. details: %s", reason))
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

func createImagePath(ownerID, name string) string {
	return fmt.Sprintf("%s/%s/%s", config.Config.Registry.URL, subsystemutils.GetPrefixedName(ownerID), name)
}

func createTransferCode() string {
	return utils.HashString(uuid.NewString())
}
