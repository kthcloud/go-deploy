package deployment_service

import (
	"fmt"
	"go-deploy/models/dto/body"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/service/deployment_service/internal_service"
	"log"
)

func Create(deploymentID, ownerID string, deploymentCreate *body.DeploymentCreate) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create deployment. details: %s", err)
	}

	params := &deploymentModel.CreateParams{}
	params.FromDTO(deploymentCreate)

	err := deploymentModel.CreateDeployment(deploymentID, ownerID, params)
	if err != nil {
		return makeError(err)
	}

	err = internal_service.CreateHarbor(params.Name, ownerID)
	if err != nil {
		return makeError(err)
	}

	_, err = internal_service.CreateK8s(params.Name, ownerID, params.Envs)
	if err != nil {
		return makeError(err)
	}

	if params.GitHub != nil {
		err = internal_service.CreateGitHub(params.Name, params)
		if err != nil {
			return makeError(err)
		}
	} else {
		err = internal_service.CreatePlaceholderGitHub(params.Name)
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
		return deployment.Ready(), "It is not ready"
	case deploymentModel.ActivityBuilding:
		return deployment.Ready(), "It is not ready"
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

	started, reason, err := StartActivity(deployment.ID, deploymentModel.ActivityBuilding)
	if err != nil {
		return makeError(err)
	}

	if !started {
		return fmt.Errorf("failed to build deployment. details: %s", reason)
	}

	defer func() {
		err = deploymentModel.RemoveActivity(deployment.ID, deploymentModel.ActivityBuilding)
		if err != nil {
			log.Println("failed to remove activity", deploymentModel.ActivityBuilding, "from deployment", deployment.Name, "details:", err)
		}
	}()

	err = internal_service.CreateBuild(deployment.ID, params)
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
