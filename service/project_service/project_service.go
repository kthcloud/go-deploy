package project_service

import (
	"deploy-api-go/models"
	"deploy-api-go/pkg/subsystems/harbor"
	"deploy-api-go/pkg/subsystems/k8s"
	"deploy-api-go/pkg/subsystems/npm"
	"go.mongodb.org/mongo-driver/bson"
	"log"
)

func Create(projectID, name, owner string) {

	go func() {
		err := models.CreateProject(projectID, name, owner)
		if err != nil {
			log.Println(err)
		}

		err = harbor.Create(name)
		if err != nil {
			log.Println(err)
		}

		err = npm.Create(name)
		if err != nil {
			log.Println(err)
		}

		err = k8s.Create(name)
		if err != nil {
			log.Println(err)
		}
	}()
}

func Get(userId, projectID string) (*models.Project, error) {
	return models.GetProjectWithOwner(userId, projectID)
}

func GetByName(userId, name string) (*models.Project, error) {
	return models.GetProjectByName(userId, name)
}

func GetByOwner(owner string) ([]models.Project, error) {
	return models.GetProjects(owner)
}

func GetAll() ([]models.Project, error) {
	return models.GetAllProjects()
}

func Exists(name string) (bool, *models.Project, error) {
	return models.ProjectExists(name)
}

func MarkBeingCreated(projectID string) error {
	return models.UpdateProject(projectID, bson.D{{
		"beingCreated", true,
	}})
}

func MarkBeingDeleted(projectID string) error {
	return models.UpdateProject(projectID, bson.D{{
		"beingDeleted", true,
	}})
}

func Delete(name string) error {
	err := harbor.Delete(name)
	if err != nil {
		log.Println(err)
		return err
	}

	err = npm.Delete(name)
	if err != nil {
		log.Println(err)
		return err
	}

	err = k8s.Delete(name)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
