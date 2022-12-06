package models

import (
	"context"
	"deploy-api-go/models/dto"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
)

type Project struct {
	ID           string    `bson:"id"`
	Name         string    `bson:"name"`
	Owner        string    `bson:"owner"`
	BeingCreated bool      `bson:"beingCreated"`
	BeingDeleted bool      `bson:"beingDeleted"`
	Subsytems    Subsystem `bson:"subsystems"`
}

type Subsystem struct {
	K8s    SubsystemK8s    `bson:"k8s"`
	Npm    SubsystemNpm    `bson:"npm"`
	Harbor SubsystemHarbor `bson:"harbor"`
}

type SubsystemK8s struct {
}

type SubsystemNpm struct {
}

type SubsystemHarbor struct {
	RobotUsername string `bson:"robotUsername"`
	RobotPassword string `bson:"robotPassword"`
}

func (p *Project) ToDto() dto.Project {
	return dto.Project{
		ID:    p.ID,
		Name:  p.Name,
		Owner: p.Owner,
	}
}

func CreateProject(projectID, name, owner string) error {
	currentProject, err := GetProjectByID(projectID)
	if err != nil {
		return err
	}

	if currentProject != nil {
		return nil
	}

	project := Project{
		ID:           projectID,
		Name:         name,
		Owner:        owner,
		BeingCreated: true,
		BeingDeleted: false,
	}

	_, err = ProjectCollection.InsertOne(context.TODO(), project)
	if err != nil {

		err = fmt.Errorf("failed to add project %s. details: %s", name, err)
		log.Println(err)
		return err
	}

	return nil
}

func getProject(filter bson.D) (*Project, error) {
	var project Project
	err := ProjectCollection.FindOne(context.TODO(), filter).Decode(&project)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}

		err = fmt.Errorf("failed to fetch project. details: %s", err)
		log.Println(err)
		invalidProject := Project{}
		return &invalidProject, err
	}

	return &project, err
}

func GetProjectByID(projectID string) (*Project, error) {
	return getProject(bson.D{{"id", projectID}})
}

func GetProjectByName(userId, name string) (*Project, error) {
	return getProject(bson.D{{"owner", userId}, {"name", name}})
}

func ProjectExists(name string) (bool, *Project, error) {
	project, err := getProject(bson.D{{"name", name}})
	if err != nil {
		return false, nil, err
	}

	if project == nil {
		return false, nil, nil
	}

	return true, project, err
}

func GetProjects(owner string) ([]Project, error) {
	cursor, err := ProjectCollection.Find(context.TODO(), bson.D{{"owner", owner}})

	if err != nil {
		err = fmt.Errorf("failed to find projects from owner %s. details: %s", owner, err)
		log.Println(err)
		return nil, err
	}

	var projects []Project
	for cursor.Next(context.TODO()) {
		var project Project

		err = cursor.Decode(&project)
		if err != nil {
			err = fmt.Errorf("failed to fetch project when fetching all project from owner %s. details: %s", owner, err)
			log.Println(err)
			return nil, err
		}
		projects = append(projects, project)
	}

	return projects, nil
}

func DeleteProject(projectId, userId string) error {
	_, err := ProjectCollection.DeleteOne(context.TODO(), bson.D{{"id", projectId}, {"owner", userId}})
	if err != nil {
		err = fmt.Errorf("failed to delete project %s. details: %s", projectId, err)
		log.Println(err)
		return err
	}
	return nil
}

func UpdateProject(id string, update bson.D) error {
	_, err := ProjectCollection.UpdateOne(context.TODO(), bson.D{{"id", id}}, bson.D{{"$set", update}})
	if err != nil {
		err = fmt.Errorf("failed to update project %s. details: %s", id, err)
		log.Println(err)
		return err
	}
	return nil
}

func UpdateProjectByName(name string, update bson.D) error {
	_, err := ProjectCollection.UpdateOne(context.TODO(), bson.D{{"name", name}}, bson.D{{"$set", update}})
	if err != nil {
		err = fmt.Errorf("failed to update project %s. details: %s", name, err)
		log.Println(err)
		return err
	}
	return nil
}

func GetAllProjects() ([]Project, error) {
	return GetAllProjectsWithCondition(bson.D{})
}

func GetAllProjectsWithCondition(condition bson.D) ([]Project, error) {
	cursor, err := ProjectCollection.Find(context.TODO(), condition)

	if err != nil {
		err = fmt.Errorf("failed to fetch all projects. details: %s", err)
		log.Println(err)
		return nil, err
	}

	var projects []Project
	for cursor.Next(context.TODO()) {
		var project Project

		err = cursor.Decode(&project)
		if err != nil {
			err = fmt.Errorf("failed to decode project when fetching all project. details: %s", err)
			log.Println(err)
			return nil, err
		}
		projects = append(projects, project)
	}

	return projects, nil
}
