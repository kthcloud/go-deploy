package deployment_service

import (
	deploymentModel "go-deploy/models/deployment"
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

		err = CreateHarbor(name)
		if err != nil {
			log.Println(err)
			return
		}

		k8sResult, err := CreateK8s(name)
		if err != nil {
			log.Println(err)
			return
		}

		err = CreateNPM(name, k8sResult.Service.GetHostName())
		if err != nil {
			log.Println(err)
			return
		}

	}()
}

func GetByID(userId, deploymentID string) (*deploymentModel.Deployment, error) {
	deployment, err := deploymentModel.GetDeploymentByID(deploymentID)
	if err != nil {
		return nil, err
	}

	if deployment != nil && deployment.Owner != userId {
		return nil, nil
	}

	return deployment, nil
}

func GetByName(userId, name string) (*deploymentModel.Deployment, error) {
	deployment, err := deploymentModel.GetDeploymentByName(name)
	if err != nil {
		return nil, err
	}

	if deployment != nil && deployment.Owner != userId {
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
		err := DeleteHarbor(name)
		if err != nil {
			log.Println(err)
			return
		}

		err = DeleteNPM(name)
		if err != nil {
			log.Println(err)
			return
		}

		err = DeleteK8s(name)
		if err != nil {
			log.Println(err)
			return
		}
	}()
}

func Restart(name string) error {
	return RestartK8s(name)
}
