package project_service

import (
	"go-deploy/models"
	"go-deploy/pkg/subsystems/harbor"
	"go-deploy/pkg/subsystems/k8s"
	"go-deploy/pkg/subsystems/npm"
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
			return
		}

		err = npm.Create(name)
		if err != nil {
			log.Println(err)
			return
		}

		err = k8s.Create(name)
		if err != nil {
			log.Println(err)
			return
		}
	}()
}

func Get(userId, projectID string) (*models.Project, error) {
	project, err := models.GetProjectByID(projectID)
	if err != nil {
		return nil, err
	}

	if project != nil && project.Owner != userId {
		return nil, nil
	}

	return project, nil
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

func MarkBeingDeleted(projectID string) error {
	return models.UpdateProject(projectID, bson.D{{
		"beingDeleted", true,
	}})
}

func Delete(name string) {
	go func() {
		err := harbor.Delete(name)
		if err != nil {
			log.Println(err)
			return
		}

		err = npm.Delete(name)
		if err != nil {
			log.Println(err)
			return
		}

		err = k8s.Delete(name)
		if err != nil {
			log.Println(err)
			return
		}
	}()
}

func Restart(name string) error {
	return k8s.Restart(name)
}
