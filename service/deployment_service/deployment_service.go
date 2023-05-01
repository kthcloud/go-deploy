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

		err = internal_service.CreateHarbor(name, owner)
		if err != nil {
			log.Println(err)
			return
		}

		k8sResult, err := internal_service.CreateK8s(name, owner)
		if err != nil {
			log.Println(err)
			return
		}

		err = internal_service.CreateNPM(name, k8sResult.Service.GetFQDN())
		if err != nil {
			log.Println(err)
			return
		}

	}()
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
	return deploymentModel.UpdateByID(deploymentID, bson.D{{
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
