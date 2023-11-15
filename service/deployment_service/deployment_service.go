package deployment_service

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/query"
	deploymentModel "go-deploy/models/sys/deployment"
	jobModel "go-deploy/models/sys/job"
	notificationModel "go-deploy/models/sys/notification"
	roleModel "go-deploy/models/sys/role"
	teamModels "go-deploy/models/sys/team"
	"go-deploy/pkg/config"
	"go-deploy/service"
	"go-deploy/service/deployment_service/base"
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

func Create(deploymentID, ownerID string, deploymentCreate *body.DeploymentCreate) error {
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
			return NonUniqueFieldErr
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
	deployment, err = deploymentModel.New().GetByID(deploymentID)
	if err != nil {
		return makeError(err)
	}

	if deployment == nil {
		return makeError(fmt.Errorf("deployment not found after creation"))
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

		err = build([]string{deployment.ID}, &deploymentModel.BuildParams{
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

func Update(id string, deploymentUpdate *body.DeploymentUpdate) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update deployment. details: %w", err)
	}

	deployment, err := deploymentModel.New().GetByID(id)
	if err != nil {
		return makeError(err)
	}

	if deployment == nil {
		log.Println("deployment", id, "not found when updating. assuming it was deleted")
		return nil
	}

	params := &deploymentModel.UpdateParams{}
	params.FromDTO(deploymentUpdate, deployment.Type)

	if params.Name != nil && deployment.Type == deploymentModel.TypeCustom {
		image := createImagePath(deployment.OwnerID, *params.Name)
		params.Image = &image
	}

	err = deploymentModel.New().UpdateWithParamsByID(id, params)
	if err != nil {
		if errors.Is(err, deploymentModel.NonUniqueFieldErr) {
			return NonUniqueFieldErr
		}

		return makeError(err)
	}

	if deployment.Type == deploymentModel.TypeCustom {
		err = harbor_service.Update(id, params)
		if err != nil {
			return makeError(err)
		}
	}

	err = k8s_service.Update(id, params)
	if err != nil {
		if errors.Is(err, base.CustomDomainInUseErr) {
			log.Println("custom domain in use when updating deployment", deployment.Name, ". removing it from the update params")
			deploymentUpdate.CustomDomain = nil
		} else {
			return makeError(err)
		}
	}

	return nil
}

func UpdateOwnerAuth(id string, params *body.DeploymentUpdateOwner, auth *service.AuthInfo) (*string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update deployment owner. details: %w", err)
	}

	deployment, err := deploymentModel.New().GetByID(id)
	if err != nil {
		return nil, makeError(err)
	}

	if deployment == nil {
		return nil, DeploymentNotFoundErr
	}

	if deployment.OwnerID == params.NewOwnerID {
		return nil, nil
	}

	doTransfer := false

	if auth.IsAdmin {
		doTransfer = true
	} else if auth.UserID == params.NewOwnerID {
		if params.TransferCode == nil || deployment.Transfer == nil || deployment.Transfer.Code != *params.TransferCode {
			return nil, InvalidTransferCodeErr
		}

		doTransfer = true
	}

	if doTransfer {
		jobID := uuid.New().String()
		err := job_service.Create(jobID, auth.UserID, jobModel.TypeUpdateDeploymentOwner, map[string]interface{}{
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
			"id":     deployment.ID,
			"name":   deployment.Name,
			"userId": params.OldOwnerID,
			"email":  auth.GetEmail(),
			"code":   code,
		},
	})

	if err != nil {
		return nil, makeError(err)
	}

	return nil, nil
}

func UpdateOwner(id string, params *body.DeploymentUpdateOwner) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update deployment owner. details: %w", err)
	}

	deployment, err := deploymentModel.New().GetByID(id)
	if err != nil {
		return makeError(err)
	}

	if deployment == nil {
		log.Println("deployment", id, "not found when updating owner. assuming it was deleted")
		return nil
	}

	var newImage *string
	if deployment.Type == deploymentModel.TypeCustom {
		image := createImagePath(params.NewOwnerID, deployment.Name)
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

	err = harbor_service.EnsureOwner(id, params.OldOwnerID)
	if err != nil {
		return makeError(err)
	}

	err = k8s_service.EnsureOwner(id, params.OldOwnerID)
	if err != nil {
		return makeError(err)
	}

	log.Println("deployment", id, "owner updated to", params.NewOwnerID)
	return nil
}

func Delete(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete deployment. details: %w", err)
	}

	err := harbor_service.Delete(id)
	if err != nil {
		return makeError(err)
	}

	err = k8s_service.Delete(id)
	if err != nil {
		return makeError(err)
	}

	err = github_service.Delete(id)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func Repair(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to repair deployment %s. details: %w", id, err)
	}

	deployment, err := deploymentModel.New().GetByID(id)
	if err != nil {
		return makeError(err)
	}

	if deployment == nil {
		log.Println("deployment", id, "not found when repairing. assuming it was deleted")
		return nil
	}

	if !deployment.Ready() {
		log.Println("deployment", id, "not ready when repairing.")
		return nil
	}

	err = k8s_service.Repair(deployment.ID)
	if err != nil {
		if errors.Is(err, base.CustomDomainInUseErr) {
			log.Println("custom domain in use when repairing deployment", deployment.ID, ". removing it from the deployment")
			err = deploymentModel.New().RemoveCustomDomain(deployment.ID)
			if err != nil {
				return makeError(err)
			}
		} else {
			return makeError(err)
		}
	}

	if !deployment.Subsystems.Harbor.Placeholder {
		err = harbor_service.Repair(deployment.ID)
		if err != nil {
			return makeError(err)
		}
	}

	log.Println("successfully repaired deployment", deployment.ID)
	return nil
}

func Restart(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to restart deployment. details: %w", err)
	}

	client := deploymentModel.New()

	deployment, err := client.GetByID(id)
	if err != nil {
		return makeError(err)
	}

	if deployment == nil {
		log.Println("deployment", id, "not found when restarting. assuming it was deleted")
		return nil
	}

	err = client.SetWithBsonByID(id, bson.D{{"restartedAt", time.Now()}})
	if err != nil {
		return makeError(err)
	}

	started, reason, err := StartActivity(deployment.ID, deploymentModel.ActivityRestarting)
	if err != nil {
		return makeError(err)
	}

	go func() {
		// the restart is best-effort, so we mimic a reasonable delay
		time.Sleep(5 * time.Second)

		err = deploymentModel.New().RemoveActivity(deployment.ID, deploymentModel.ActivityRestarting)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to remove activity %s from deployment %s. details: %w", deploymentModel.ActivityRestarting, deployment.ID, err))
		}
	}()

	if !started {
		return fmt.Errorf("failed to restart deployment. details: %s", reason)
	}

	err = k8s_service.Restart(id)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func Build(ids []string, buildParams *body.DeploymentBuild) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to build deployment. details: %w", err)
	}

	params := &deploymentModel.BuildParams{}
	params.FromDTO(buildParams)

	err := build(ids, params)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func DoCommand(deployment *deploymentModel.Deployment, command string) {
	go func() {
		switch command {
		case "restart":
			err := Restart(deployment.ID)
			if err != nil {
				utils.PrettyPrintError(err)
			}
		}
	}()
}

func GetByIdAuth(id string, auth *service.AuthInfo) (*deploymentModel.Deployment, error) {
	deployment, err := deploymentModel.New().GetByID(id)
	if err != nil {
		return nil, err
	}

	if deployment == nil {
		return nil, nil
	}

	if deployment.OwnerID != auth.UserID {
		inTeam, err := teamModels.New().AddUserID(auth.UserID).AddResourceID(id).ExistsAny()
		if err != nil {
			return nil, err
		}

		if inTeam {
			return deployment, nil
		}

		if auth.IsAdmin {
			return deployment, nil
		}

		return nil, nil
	}

	return deployment, nil
}

func GetByID(id string) (*deploymentModel.Deployment, error) {
	return deploymentModel.New().GetByID(id)
}

func GetByName(name string) (*deploymentModel.Deployment, error) {
	return deploymentModel.New().GetByName(name)
}

func GetByTransferCode(code, userID string) (*deploymentModel.Deployment, error) {
	return deploymentModel.New().GetByTransferCode(code, userID)
}

func NameAvailable(name string) (bool, error) {
	exists, err := deploymentModel.New().ExistsByName(name)
	if err != nil {
		return false, err
	}

	return !exists, nil
}

func ListAuth(allUsers bool, userID *string, shared bool, auth *service.AuthInfo, pagination *query.Pagination) ([]deploymentModel.Deployment, error) {
	client := deploymentModel.New()

	if pagination != nil {
		client.AddPagination(pagination.Page, pagination.PageSize)
	}

	if userID != nil {
		if *userID != auth.UserID && !auth.IsAdmin {
			return nil, nil
		}
		client.RestrictToOwner(*userID)
	} else if !allUsers || (allUsers && !auth.IsAdmin) {
		client.RestrictToOwner(auth.UserID)
	}

	resources, err := client.ListAll()
	if err != nil {
		return nil, err
	}

	if shared {
		ids := make([]string, len(resources))
		for i, resource := range resources {
			ids[i] = resource.ID
		}

		teamClient := teamModels.New().AddUserID(auth.UserID)
		if pagination != nil {
			teamClient.AddPagination(pagination.Page, pagination.PageSize)
		}

		teams, err := teamClient.ListAll()
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
				for _, id := range ids {
					if resource.ID == id {
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
		if pagination != nil {
			resources = utils.GetPage(resources, pagination.PageSize, pagination.Page)
		}

	} else {
		// sort by createdAt
		sort.Slice(resources, func(i, j int) bool {
			return resources[i].CreatedAt.After(resources[j].CreatedAt)
		})
	}

	return resources, nil
}

func ListAll() ([]deploymentModel.Deployment, error) {
	return deploymentModel.New().ListAll()
}

func CheckQuotaCreate(userID string, quota *roleModel.Quotas, auth *service.AuthInfo) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check quota. details: %w", err)
	}

	if auth.IsAdmin {
		return nil
	}

	usage, err := GetUsageByUserID(userID)
	if err != nil {
		return makeError(err)
	}

	totalCount := usage.Count + 1

	if totalCount > quota.Deployments {
		return service.NewQuotaExceededError(fmt.Sprintf("Deployment quota exceeded. Current: %d, Quota: %d", totalCount, quota.CpuCores))
	}

	return nil
}

func StartActivity(deploymentID, activity string) (bool, string, error) {
	canAdd, reason := CanAddActivity(deploymentID, activity)
	if !canAdd {
		return false, reason, nil
	}

	err := deploymentModel.New().AddActivity(deploymentID, activity)
	if err != nil {
		return false, "", err
	}

	return true, "", nil
}

func CanAddActivity(deploymentID, activity string) (bool, string) {
	deployment, err := deploymentModel.New().GetByID(deploymentID)
	if err != nil {
		return false, "Failed to get deployment"
	}

	if deployment == nil {
		return false, "Deployment not found"
	}

	switch activity {
	case deploymentModel.ActivityBeingCreated:
		return !deployment.BeingDeleted(), "Resource is being deleted"
	case deploymentModel.ActivityBeingDeleted:
		return true, ""
	case deploymentModel.ActivityUpdating:
		return !deployment.BeingDeleted() && !deployment.BeingCreated(), "Resource is being deleted or created"
	case deploymentModel.ActivityRestarting:
		return !deployment.BeingDeleted(), "Resource is being deleted"
	case deploymentModel.ActivityBuilding:
		return !deployment.BeingDeleted(), "Resource is being deleted"
	case deploymentModel.ActivityRepairing:
		return deployment.Ready(), "Resource is not ready"
	}

	return false, fmt.Sprintf("Unknown activity %s", activity)
}

func GetUsageByUserID(userID string) (*deploymentModel.Usage, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get usage. details: %w", err)
	}

	count, err := deploymentModel.New().RestrictToOwner(userID).Count()
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

func build(ids []string, params *deploymentModel.BuildParams) error {
	var filtered []string
	for _, id := range ids {
		started, reason, err := StartActivity(id, deploymentModel.ActivityBuilding)
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
