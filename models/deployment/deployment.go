package deployment

import (
	"context"
	"fmt"
	"go-deploy/models"
	"go-deploy/models/dto"
	harborModels "go-deploy/pkg/subsystems/harbor/models"
	npmModels "go-deploy/pkg/subsystems/npm/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
)

type Deployment struct {
	ID           string               `bson:"id"`
	Name         string               `bson:"name"`
	Owner        string               `bson:"owner"`
	BeingCreated bool                 `bson:"beingCreated"`
	BeingDeleted bool                 `bson:"beingDeleted"`
	Subsystems   DeploymentSubsystems `bson:"subsystems"`
}

type DeploymentSubsystems struct {
	K8s    DeploymentK8s    `bson:"k8s"`
	Npm    DeploymentNPM    `bson:"npm"`
	Harbor DeploymentHarbor `bson:"harbor"`
}

type DeploymentK8s struct {
}

type DeploymentNPM struct {
	ProxyHost npmModels.ProxyHostPublic `bson:"proxyHost"`
}

type DeploymentHarbor struct {
	Project    harborModels.ProjectPublic    `bson:"project"`
	Robot      harborModels.RobotPublic      `bson:"robot"`
	Repository harborModels.RepositoryPublic `bson:"repository"`
	Webhook    harborModels.WebhookPublic    `bson:"webhook"`
}

func (deployment *Deployment) ToDto() dto.DeploymentRead {
	return dto.DeploymentRead{
		ID:    deployment.ID,
		Name:  deployment.Name,
		Owner: deployment.Owner,
	}
}

func (deployment *Deployment) Ready() bool {
	return !deployment.BeingCreated && !deployment.BeingDeleted
}

func CreateDeployment(deploymentID, name, owner string) error {
	currentDeployment, err := GetDeploymentByID(deploymentID)
	if err != nil {
		return err
	}

	if currentDeployment != nil {
		return nil
	}

	deployment := Deployment{
		ID:           deploymentID,
		Name:         name,
		Owner:        owner,
		BeingCreated: true,
		BeingDeleted: false,
	}

	_, err = models.DeploymentCollection.InsertOne(context.TODO(), deployment)
	if err != nil {
		err = fmt.Errorf("failed to create deployment %s. details: %s", name, err)
		return err
	}

	return nil
}

func getDeployment(filter bson.D) (*Deployment, error) {
	var deployment Deployment
	err := models.DeploymentCollection.FindOne(context.TODO(), filter).Decode(&deployment)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}

		err = fmt.Errorf("failed to fetch deployment. details: %s", err)
		invalidDeployment := Deployment{}
		return &invalidDeployment, err
	}

	return &deployment, err
}

func GetDeploymentByID(deploymentID string) (*Deployment, error) {
	return getDeployment(bson.D{{"id", deploymentID}})
}

func GetDeploymentByName(name string) (*Deployment, error) {
	return getDeployment(bson.D{{"name", name}})
}

func GetByWebhookToken(token string) (*Deployment, error) {
	return getDeployment(bson.D{{"subsystems.harbor.webhook.token", token}})
}

func DeploymentExists(name string) (bool, *Deployment, error) {
	deployment, err := getDeployment(bson.D{{"name", name}})
	if err != nil {
		return false, nil, err
	}

	if deployment == nil {
		return false, nil, nil
	}

	return true, deployment, err
}

func GetDeployments(owner string) ([]Deployment, error) {
	cursor, err := models.DeploymentCollection.Find(context.TODO(), bson.D{{"owner", owner}})

	if err != nil {
		err = fmt.Errorf("failed to find deployments from owner %s. details: %s", owner, err)
		log.Println(err)
		return nil, err
	}

	var deployments []Deployment
	for cursor.Next(context.TODO()) {
		var deployment Deployment

		err = cursor.Decode(&deployment)
		if err != nil {
			err = fmt.Errorf("failed to fetch deployment when fetching all deployments from owner %s. details: %s", owner, err)
			log.Println(err)
			return nil, err
		}
		deployments = append(deployments, deployment)
	}

	return deployments, nil
}

func DeleteDeployment(deploymentID, userId string) error {
	_, err := models.DeploymentCollection.DeleteOne(context.TODO(), bson.D{{"id", deploymentID}, {"owner", userId}})
	if err != nil {
		err = fmt.Errorf("failed to delete deployment %s. details: %s", deploymentID, err)
		log.Println(err)
		return err
	}
	return nil
}

func UpdateDeployment(id string, update bson.D) error {
	_, err := models.DeploymentCollection.UpdateOne(context.TODO(), bson.D{{"id", id}}, bson.D{{"$set", update}})
	if err != nil {
		err = fmt.Errorf("failed to update deployment %s. details: %s", id, err)
		log.Println(err)
		return err
	}
	return nil
}

func UpdateDeploymentByName(name string, update bson.D) error {
	_, err := models.DeploymentCollection.UpdateOne(context.TODO(), bson.D{{"name", name}}, bson.D{{"$set", update}})
	if err != nil {
		err = fmt.Errorf("failed to update deployment %s. details: %s", name, err)
		log.Println(err)
		return err
	}
	return nil
}

func UpdateSubsystemByName(name, subsystem string, key string, update interface{}) error {
	subsystemKey := fmt.Sprintf("subsystems.%s.%s", subsystem, key)
	return UpdateDeploymentByName(name, bson.D{{subsystemKey, update}})
}

func GetAllDeployments() ([]Deployment, error) {
	return GetAllDeploymentsWithFilter(bson.D{})
}

func GetAllDeploymentsWithFilter(filter bson.D) ([]Deployment, error) {
	cursor, err := models.DeploymentCollection.Find(context.TODO(), filter)

	if err != nil {
		err = fmt.Errorf("failed to fetch all deployments. details: %s", err)
		log.Println(err)
		return nil, err
	}

	var deployments []Deployment
	for cursor.Next(context.TODO()) {
		var deployment Deployment

		err = cursor.Decode(&deployment)
		if err != nil {
			err = fmt.Errorf("failed to decode deployment when fetching all deployment. details: %s", err)
			log.Println(err)
			return nil, err
		}
		deployments = append(deployments, deployment)
	}

	return deployments, nil
}
