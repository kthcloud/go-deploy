package deployment_service

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/query"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/models/sys/deployment/storage_manager"
	roleModel "go-deploy/models/sys/enviroment/role"
	jobModel "go-deploy/models/sys/job"
	"go-deploy/pkg/conf"
	"go-deploy/service"
	"go-deploy/service/deployment_service/github_service"
	"go-deploy/service/deployment_service/gitlab_service"
	"go-deploy/service/deployment_service/harbor_service"
	"go-deploy/service/deployment_service/k8s_service"
	"go-deploy/service/job_service"
	"go-deploy/utils"
	"go-deploy/utils/subsystemutils"
	"log"
	"strings"
)

func Create(deploymentID, ownerID string, deploymentCreate *body.DeploymentCreate) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create deployment. details: %w", err)
	}

	// temporary hard-coded fallback
	fallbackZone := "se-flem"
	fallbackImage := fmt.Sprintf("%s/%s/%s", conf.Env.DockerRegistry.URL, subsystemutils.GetPrefixedName(ownerID), deploymentCreate.Name)
	fallbackPort := conf.Env.Deployment.Port

	params := &deploymentModel.CreateParams{}
	params.FromDTO(deploymentCreate, fallbackZone, fallbackImage, fallbackPort)

	// create storage manager only if needed
	if len(params.Volumes) > 0 {
		storageManagerID := uuid.New().String()
		jobID := uuid.New().String()
		err := job_service.Create(jobID, ownerID, jobModel.TypeCreateStorageManager, map[string]interface{}{
			"id": storageManagerID,
			"params": storage_manager.CreateParams{
				UserID: ownerID,
				Zone:   params.Zone,
			},
		})

		if err != nil {
			return makeError(err)
		}
	}

	deployment, err := deploymentModel.New().Create(deploymentID, ownerID, params)
	if err != nil {
		return makeError(err)
	}

	if deployment == nil {
		return makeError(fmt.Errorf("deployment already exists for another user"))
	}

	if deployment.Type == deploymentModel.TypeCustom {
		err = harbor_service.Create(deploymentID, ownerID, params)
		if err != nil {
			return makeError(err)
		}
	} else {
		err = harbor_service.CreatePlaceholder(params.Name)
		if err != nil {
			return makeError(err)
		}
	}

	err = k8s_service.Create(deploymentID, ownerID, params)
	if err != nil {
		if errors.Is(err, k8s_service.CustomDomainInUseErr) {
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
		err = github_service.CreatePlaceholder(params.Name)
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

		err = build(deployment, &deploymentModel.BuildParams{
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

func GetByIdAuth(deploymentID string, auth *service.AuthInfo) (*deploymentModel.Deployment, error) {
	deployment, err := deploymentModel.New().GetByID(deploymentID)
	if err != nil {
		return nil, err
	}

	if deployment == nil {
		return nil, nil
	}

	if deployment.OwnerID != auth.UserID && !auth.IsAdmin {
		return nil, nil
	}

	return deployment, nil
}

func GetManyAuth(allUsers bool, userID *string, auth *service.AuthInfo, pagination *query.Pagination) ([]deploymentModel.Deployment, error) {
	client := deploymentModel.New()

	if pagination != nil {
		client.AddPagination(pagination.Page, pagination.PageSize)
	}

	if userID != nil {
		if *userID != auth.UserID && !auth.IsAdmin {
			return nil, nil
		}
		client.RestrictToUser(userID)
	} else if !allUsers || (allUsers && !auth.IsAdmin) {
		client.RestrictToUser(&auth.UserID)
	}

	return client.GetMany()
}

func GetAll() ([]deploymentModel.Deployment, error) {
	return deploymentModel.New().GetAll()
}

func GetCountAuth(userID string, auth *service.AuthInfo) (int, error) {
	if userID != auth.UserID && !auth.IsAdmin {
		return 0, nil
	}

	return deploymentModel.New().CountByOwnerID(userID)
}

func CheckQuotaCreate(userID string, quota *roleModel.Quotas, auth *service.AuthInfo) (bool, string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check quota. details: %w", err)
	}

	if auth.IsAdmin {
		return true, "", nil
	}

	usage, err := GetUsageByUserID(userID)
	if err != nil {
		return false, "", makeError(err)
	}

	totalCount := usage.Count + 1

	if totalCount > quota.Deployments {
		return false, fmt.Sprintf("Deployment quota exceeded. Current: %d, Quota: %d", totalCount, quota.CpuCores), nil
	}

	return true, "", nil
}

func GetByName(name string) (*deploymentModel.Deployment, error) {
	return deploymentModel.New().GetByName(name)
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
		return !deployment.BeingDeleted(), "It is being deleted"
	case deploymentModel.ActivityBeingDeleted:
		return true, ""
	case deploymentModel.ActivityRestarting:
		return !deployment.BeingDeleted(), "It is being deleted"
	case deploymentModel.ActivityBuilding:
		return !deployment.BeingDeleted(), "It is being deleted"
	case deploymentModel.ActivityRepairing:
		return deployment.Ready(), "It is not ready"
	}

	return false, fmt.Sprintf("Unknown activity %s", activity)
}

func Delete(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete deployment. details: %w", err)
	}

	err := harbor_service.Delete(name)
	if err != nil {
		return makeError(err)
	}

	err = k8s_service.Delete(name)
	if err != nil {
		return makeError(err)
	}

	err = github_service.Delete(name, nil)
	if err != nil {
		return makeError(err)
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

	err = k8s_service.Update(deployment.Name, params)
	if err != nil {
		if errors.Is(err, k8s_service.CustomDomainInUseErr) {
			log.Println("custom domain in use when updating deployment", deployment.Name, ". removing it from the update params")
			deploymentUpdate.CustomDomain = nil
		} else {
			return makeError(err)
		}
	}

	err = deploymentModel.New().UpdateWithParamsByID(id, params)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func Restart(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to restart deployment. details: %w", err)
	}

	deployment, err := deploymentModel.New().GetByName(name)
	if err != nil {
		return makeError(err)
	}

	if deployment == nil {
		log.Println("deployment", name, "not found when restarting. assuming it was deleted")
		return nil
	}

	started, reason, err := StartActivity(deployment.ID, deploymentModel.ActivityRestarting)
	if err != nil {
		return makeError(err)
	}

	if !started {
		return fmt.Errorf("failed to restart deployment. details: %s", reason)
	}

	err = k8s_service.Restart(name)
	if err != nil {
		return makeError(err)
	}

	err = deploymentModel.New().RemoveActivity(deployment.ID, deploymentModel.ActivityRestarting)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func Build(id string, buildParams *body.DeploymentBuild) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to build deployment. details: %w", err)
	}

	deployment, err := deploymentModel.New().GetByID(id)
	if err != nil {
		return makeError(err)
	}

	if deployment == nil {
		log.Println("deployment", id, "not found when building. assuming it was deleted")
		return nil
	}

	params := &deploymentModel.BuildParams{}
	params.FromDTO(buildParams)

	err = build(deployment, params)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func DoCommand(vm *deploymentModel.Deployment, command string) {
	go func() {
		switch command {
		case "restart":
			_ = Restart(vm.Name)
		}
	}()
}

func GetUsageByUserID(userID string) (*deploymentModel.Usage, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get usage. details: %w", err)
	}

	count, err := deploymentModel.New().CountByOwnerID(userID)
	if err != nil {
		return nil, makeError(err)
	}

	return &deploymentModel.Usage{
		Count: count,
	}, nil
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

	started, reason, err := StartActivity(deployment.ID, deploymentModel.ActivityRepairing)
	if err != nil {
		return makeError(err)
	}

	if !started {
		return fmt.Errorf("failed to repair deployment. details: %s", reason)
	}

	defer func() {
		err = deploymentModel.New().RemoveActivity(deployment.ID, deploymentModel.ActivityRepairing)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to remove activity %s from deployment %s. details: %w", deploymentModel.ActivityRepairing, deployment.Name, err))
		}
	}()

	err = k8s_service.Repair(deployment.Name)
	if err != nil {
		if errors.Is(err, k8s_service.CustomDomainInUseErr) {
			log.Println("custom domain in use when repairing deployment", deployment.Name, ". removing it from the deployment")
			err = deploymentModel.New().RemoveCustomDomain(deployment.ID)
			if err != nil {
				return makeError(err)
			}
		} else {
			return makeError(err)
		}
	}

	if !deployment.Subsystems.Harbor.Placeholder {
		err = harbor_service.Repair(deployment.Name)
		if err != nil {
			return makeError(err)
		}
	}

	log.Println("successfully repaired deployment", deployment.Name)
	return nil
}

func ValidGitHubToken(token string) (bool, string, error) {
	return github_service.ValidateToken(token)
}

func GetGitHubAccessTokenByCode(code string) (string, error) {
	return github_service.GetAccessTokenByCode(code)
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

func build(deployment *deploymentModel.Deployment, params *deploymentModel.BuildParams) error {
	started, reason, err := StartActivity(deployment.ID, deploymentModel.ActivityBuilding)
	if err != nil {
		return err
	}

	if !started {
		return fmt.Errorf("failed to build deployment. details: %s", reason)
	}

	defer func() {
		err = deploymentModel.New().RemoveActivity(deployment.ID, deploymentModel.ActivityBuilding)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to remove activity %s for deployment %s. details: %w", deploymentModel.ActivityBuilding, deployment.Name, err))
		}
	}()

	err = gitlab_service.CreateBuild(deployment.ID, params)
	if err != nil {
		// we treat building as a non-critical activity, so we don't return an error here
		utils.PrettyPrintError(fmt.Errorf("failed to create build for deployment %s details: %w", deployment.Name, err))
	}

	return nil
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
