package deployment_service

import (
	deploymentModel "go-deploy/models/deployment"
	"go-deploy/service/deployment_service/internal_service"
	"go.mongodb.org/mongo-driver/bson"
	"log"
)

func Create(deploymentID, name, owner string) {
	go func() {
		err := deploymentModel.CreateDeployment(deploymentID, name, owner)
		if err != nil {
			log.Println(err)
			return
		}

		err = internal_service.CreateHarbor(name)
		if err != nil {
			log.Println(err)
			return
		}

		k8sResult, err := internal_service.CreateK8s(name)
		if err != nil {
			log.Println(err)
			return
		}

		err = internal_service.CreateNPM(name, k8sResult.Service.GetHostName())
		if err != nil {
			log.Println(err)
			return
		}

	}()
}

func GetByFullID(userId, deploymentID string) (*deploymentModel.Deployment, error) {
	deployment, err := deploymentModel.GetDeploymentByID(deploymentID)
	if err != nil {
		return nil, err
	}

	if deployment != nil && deployment.OwnerID != userId {
		return nil, nil
	}

	return deployment, nil
}

func GetByID(deploymentID string) (*deploymentModel.Deployment, error) {
	deployment, err := deploymentModel.GetDeploymentByID(deploymentID)
	if err != nil {
		return nil, err
	}

	return deployment, nil
}

func GetByName(userId, name string) (*deploymentModel.Deployment, error) {
	deployment, err := deploymentModel.GetDeploymentByName(name)
	if err != nil {
		return nil, err
	}

	if deployment != nil && deployment.OwnerID != userId {
		return nil, nil
	}

	return deployment, nil
}

func GetByOwnerID(owner string) ([]deploymentModel.Deployment, error) {
	return deploymentModel.GetDeployments(owner)
}

func GetAll() ([]deploymentModel.Deployment, error) {
	return deploymentModel.GetAllDeployments()
}

func Exists(name string) (bool, *deploymentModel.Deployment, error) {
	return deploymentModel.DeploymentExists(name)
}

func MarkBeingDeleted(deploymentID string) error {
	return deploymentModel.UpdateDeployment(deploymentID, bson.D{{
		"beingDeleted", true,
	}})
}

func Delete(name string) {
	go func() {
		err := internal_service.DeleteHarbor(name)
		if err != nil {
			log.Println(err)
			return
		}

		err = internal_service.DeleteNPM(name)
		if err != nil {
			log.Println(err)
			return
		}

		err = internal_service.DeleteK8s(name)
		if err != nil {
			log.Println(err)
			return
		}
	}()
}

func Restart(name string) error {
	return internal_service.RestartK8s(name)
}
