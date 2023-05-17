package deployment_service

import (
	"fmt"
	"go-deploy/models/dto/body"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/service/deployment_service/internal_service"
)

func Create(deploymentID, ownerID string, deploymentCreate *body.DeploymentCreate) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create deployment. details: %s", err)
	}

	params := deploymentModel.CreateParams{}
	params.FromDTO(deploymentCreate)

	err := deploymentModel.CreateDeployment(deploymentID, ownerID, &params)
	if err != nil {
		return makeError(err)
	}

	err = internal_service.CreateHarbor(params.Name, ownerID)
	if err != nil {
		return makeError(err)
	}

	k8sResult, err := internal_service.CreateK8s(params.Name, ownerID, params.Envs)
	if err != nil {
		return makeError(err)
	}

	if !params.Private {
		err = internal_service.CreateNPM(params.Name, k8sResult.Service.GetFQDN())
		if err != nil {
			return makeError(err)
		}
	} else {
		err := internal_service.CreateFakeNPM(params.Name)
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
		return !deployment.BeingCreated() && !deployment.BeingDeleted(), "It is not ready"
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

	err = internal_service.DeleteNPM(name)
	if err != nil {
		return makeError(err)
	}

	err = internal_service.DeleteK8s(name)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func Update(name string, deploymentUpdate *body.DeploymentUpdate) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update deployment. details: %s", err)
	}

	update := deploymentModel.UpdateParams{}
	update.FromDTO(deploymentUpdate)

	deployment, err := deploymentModel.GetByName(name)
	if err != nil {
		return makeError(err)
	}

	err = internal_service.UpdateK8s(name, update.Envs)
	if err != nil {
		return makeError(err)
	}

	if update.Private != nil {
		if *update.Private && !deployment.Private {
			err = internal_service.DeleteNPM(name)
			if err != nil {
				return makeError(err)
			}
		} else if !*update.Private && deployment.Private {
			err = internal_service.DeleteNPM(name)
			if err != nil {
				return makeError(err)
			}

			err = internal_service.CreateNPM(name, deployment.Subsystems.K8s.Service.GetFQDN())
			if err != nil {
				return makeError(err)
			}
		}
	}

	err = deploymentModel.UpdateByID(deployment.ID, &update)
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
