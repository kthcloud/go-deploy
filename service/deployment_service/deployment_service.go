package deployment_service

import (
	"fmt"
	"go-deploy/models/dto/body"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/service/deployment_service/internal_service"
	"log"
	"strings"
)

func Create(deploymentID, ownerID string, deploymentCreate *body.DeploymentCreate) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create deployment. details: %s", err)
	}

	params := &deploymentModel.CreateParams{}
	params.FromDTO(deploymentCreate)

	created, err := deploymentModel.CreateDeployment(deploymentID, ownerID, params)
	if err != nil {
		return makeError(err)
	}

	if !created {
		return makeError(fmt.Errorf("deployment already exists for another user"))
	}

	err = internal_service.CreateHarbor(deploymentID, ownerID, params)
	if err != nil {
		return makeError(err)
	}

	_, err = internal_service.CreateK8s(deploymentID, ownerID, params)
	if err != nil {
		return makeError(err)
	}

	if params.GitHub != nil {
		err = internal_service.CreateGitHub(deploymentID, params)
		if err != nil {
			errString := err.Error()
			if strings.Contains(errString, "/hooks: 404 Not Found") {
				log.Println(makeError(fmt.Errorf("webhook api not found. assuming github is not supported, inserting placeholder instead")))
				err = internal_service.CreatePlaceholderGitHub(params.Name)
				if err != nil {
					return makeError(err)
				}
			} else {
				return makeError(err)
			}
		}
	} else {
		err = internal_service.CreatePlaceholderGitHub(params.Name)
		if err != nil {
			return makeError(err)
		}
	}

	deployment, err := deploymentModel.GetByID(deploymentID)
	if err != nil {
		return makeError(err)
	}

	if deployment == nil {
		return makeError(fmt.Errorf("deployment not found after creation"))
	}

	if deployment.Subsystems.GitHub.Created() && params.GitHub != nil {
		repo, err := internal_service.GetGitHubRepository(params.GitHub.Token, params.GitHub.RepositoryID)
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

		if repo.DefaultBranch == "" {
			log.Println("github repository has no default branch. assuming it was deleted")
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

func GetByID(userId, deploymentID string, isAdmin bool) (*deploymentModel.Deployment, error) {
	deployment, err := deploymentModel.GetByID(deploymentID)
	if err != nil {
		return nil, err
	}

	if deployment != nil && deployment.OwnerID != userId && !isAdmin {
		return nil, nil
	}

	return deployment, nil
}

func GetByName(userId, name string) (*deploymentModel.Deployment, error) {
	deployment, err := deploymentModel.GetByName(name)
	if err != nil {
		return nil, err
	}

	if deployment != nil && deployment.OwnerID != userId {
		return nil, nil
	}

	return deployment, nil
}

func GetByOwnerID(owner string) ([]deploymentModel.Deployment, error) {
	return deploymentModel.GetMany(owner)
}

func GetAll() ([]deploymentModel.Deployment, error) {
	return deploymentModel.GetAll()
}

func GetCount(userID string) (int, error) {
	return deploymentModel.CountByOwnerID(userID)
}

func Exists(name string) (bool, *deploymentModel.Deployment, error) {
	return deploymentModel.Exists(name)
}

func StartActivity(deploymentID, activity string) (bool, string, error) {
	canAdd, reason := CanAddActivity(deploymentID, activity)
	if !canAdd {
		return false, reason, nil
	}

	err := deploymentModel.AddActivity(deploymentID, activity)
	if err != nil {
		return false, "", err
	}

	return true, "", nil
}

func CanAddActivity(deploymentID, activity string) (bool, string) {
	deployment, err := deploymentModel.GetByID(deploymentID)
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
		return !deployment.BeingCreated(), "It is being created"
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
		return fmt.Errorf("failed to delete deployment. details: %s", err)
	}

	err := internal_service.DeleteHarbor(name)
	if err != nil {
		return makeError(err)
	}

	err = internal_service.DeleteK8s(name)
	if err != nil {
		return makeError(err)
	}

	err = internal_service.DeleteGitHub(name, nil)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func Update(id string, deploymentUpdate *body.DeploymentUpdate) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update deployment. details: %s", err)
	}

	params := &deploymentModel.UpdateParams{}
	params.FromDTO(deploymentUpdate)

	deployment, err := deploymentModel.GetByID(id)
	if err != nil {
		return makeError(err)
	}

	if deployment == nil {
		log.Println("deployment", id, "not found when updating. assuming it was deleted")
		return nil
	}

	err = internal_service.UpdateK8s(deployment.Name, params)
	if err != nil {
		return makeError(err)
	}

	err = deploymentModel.UpdateByID(id, params)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func Restart(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to restart deployment. details: %s", err)
	}

	deployment, err := deploymentModel.GetByName(name)
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

	err = internal_service.RestartK8s(name)
	if err != nil {
		return makeError(err)
	}

	err = deploymentModel.RemoveActivity(deployment.ID, deploymentModel.ActivityRestarting)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func Build(id string, buildParams *body.DeploymentBuild) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to build deployment. details: %s", err)
	}

	deployment, err := deploymentModel.GetByID(id)
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
		return fmt.Errorf("failed to get usage. details: %s", err)
	}

	count, err := deploymentModel.CountByOwnerID(userID)
	if err != nil {
		return nil, makeError(err)
	}

	return &deploymentModel.Usage{
		Count: count,
	}, nil
}

func Repair(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to repair deployment %s. details: %s", id, err)
	}

	deployment, err := deploymentModel.GetByID(id)
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
		err = deploymentModel.RemoveActivity(deployment.ID, deploymentModel.ActivityRepairing)
		if err != nil {
			log.Println("failed to remove activity", deploymentModel.ActivityRepairing, "from deployment", deployment.Name, "details:", err)
		}
	}()

	err = internal_service.RepairK8s(deployment.Name)
	if err != nil {
		return makeError(err)
	}

	err = internal_service.RepairHarbor(deployment.Name)
	if err != nil {
		return makeError(err)
	}

	log.Println("successfully repaired deployment", deployment.Name)
	return nil
}

func ValidGitHubToken(token string) (bool, string, error) {
	return internal_service.ValidateGitHubToken(token)
}

func GetGitHubAccessTokenByCode(code string) (string, error) {
	return internal_service.GetGitHubAccessTokenByCode(code)
}

func GetGitHubRepositories(token string) ([]deploymentModel.GitHubRepository, error) {
	return internal_service.GetGitHubRepositories(token)
}

func ValidGitHubRepository(token string, repositoryID int64) (bool, string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get github repository. details: %s", err)
	}

	repo, err := internal_service.GetGitHubRepository(token, repositoryID)
	if err != nil {
		return false, "", makeError(err)
	}

	if repo == nil {
		return false, "Repository not found", nil
	}

	webhooks, err := internal_service.GetGitHubWebhooks(token, repo.Owner, repo.Name)
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
		err = deploymentModel.RemoveActivity(deployment.ID, deploymentModel.ActivityBuilding)
		if err != nil {
			log.Println("failed to remove activity", deploymentModel.ActivityBuilding, "for deployment", deployment.Name, "details:", err)
		}
	}()

	err = internal_service.CreateBuild(deployment.ID, params)
	if err != nil {
		// we treat building as a non-critical activity, so we don't return an error here
		log.Println("failed to create build for deployment", deployment.Name, "details:", err)
	}

	return nil
}
