package deployment_service

import (
	"go-deploy/models"
	"go-deploy/service/subsystem_service"
	"go.mongodb.org/mongo-driver/bson"
	"log"
)

func Create(deploymentID, name, owner string) {

	go func() {
		err := models.CreateDeployment(deploymentID, name, owner)
		if err != nil {
			log.Println(err)
		}

		err = subsystem_service.CreateHarbor(name)
		if err != nil {
			log.Println(err)
			return
		}

		err = subsystem_service.CreateNPM(name)
		if err != nil {
			log.Println(err)
			return
		}

		err = subsystem_service.CreateK8s(name)
		if err != nil {
			log.Println(err)
			return
		}
	}()
}

func Get(userId, deploymentID string) (*models.Deployment, error) {
	deployment, err := models.GetDeploymentByID(deploymentID)
	if err != nil {
		return nil, err
	}

	if deployment != nil && deployment.Owner != userId {
		return nil, nil
	}

	return deployment, nil
}

func GetByName(userId, name string) (*models.Deployment, error) {
	return models.GetDeploymentByName(userId, name)
}

func GetByOwner(owner string) ([]models.Deployment, error) {
	return models.GetDeployments(owner)
}

func GetAll() ([]models.Deployment, error) {
	return models.GetAllDeployments()
}

func Exists(name string) (bool, *models.Deployment, error) {
	return models.DeploymentExists(name)
}

func MarkBeingDeleted(deploymentID string) error {
	return models.UpdateDeployment(deploymentID, bson.D{{
		"beingDeleted", true,
	}})
}

func Delete(name string) {
	go func() {
		err := subsystem_service.DeleteHarbor(name)
		if err != nil {
			log.Println(err)
			return
		}

		err = subsystem_service.DeleteNPM(name)
		if err != nil {
			log.Println(err)
			return
		}

		err = subsystem_service.DeleteK8s(name)
		if err != nil {
			log.Println(err)
			return
		}
	}()
}

func Restart(name string) error {
	return subsystem_service.RestartK8s(name)
}
