package deployment_service

import (
	"fmt"
	"go-deploy/models/dto/body"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/service/deployment_service/internal_service"
	"go.mongodb.org/mongo-driver/bson"
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

	err = internal_service.CreateNPM(params.Name, k8sResult.Service.GetFQDN())
	if err != nil {
		return makeError(err)
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

func MarkBeingDeleted(deploymentID string) error {
	return deploymentModel.UpdateWithBsonByID(deploymentID, bson.D{{
		"beingDeleted", true,
	}})
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

	err := deploymentModel.UpdateByID(name, &update)
	if err != nil {
		return makeError(err)
	}

	err = internal_service.UpdateK8s(name, update.Envs)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func Restart(name string) error {
	return internal_service.RestartK8s(name)
}
