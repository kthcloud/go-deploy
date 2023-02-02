package deployment

import (
	"context"
	"fmt"
	"go-deploy/models"
	"go-deploy/models/dto"
	harborModels "go-deploy/pkg/subsystems/harbor/models"
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
	npmModels "go-deploy/pkg/subsystems/npm/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
)

type Deployment struct {
	ID           string     `bson:"id"`
	Name         string     `bson:"name"`
	OwnerID      string     `bson:"ownerId"`
	BeingCreated bool       `bson:"beingCreated"`
	BeingDeleted bool       `bson:"beingDeleted"`
	Subsystems   Subsystems `bson:"subsystems"`
}

type Subsystems struct {
	K8s    K8s    `bson:"k8s"`
	Npm    NPM    `bson:"npm"`
	Harbor Harbor `bson:"harbor"`
}

type K8s struct {
	Namespace  k8sModels.NamespacePublic  `bson:"namespace"`
	Deployment k8sModels.DeploymentPublic `bson:"deployment"`
	Service    k8sModels.ServicePublic    `bson:"service"`
}

type NPM struct {
	ProxyHost npmModels.ProxyHostPublic `bson:"proxyHost"`
}

type Harbor struct {
	Project    harborModels.ProjectPublic    `bson:"project"`
	Robot      harborModels.RobotPublic      `bson:"robot"`
	Repository harborModels.RepositoryPublic `bson:"repository"`
	Webhook    harborModels.WebhookPublic    `bson:"webhook"`
}

func (deployment *Deployment) ToDTO(status string, url string) dto.DeploymentRead {
	fullURL := fmt.Sprintf("https://%s", url)
	return dto.DeploymentRead{
		ID:      deployment.ID,
		Name:    deployment.Name,
		OwnerID: deployment.OwnerID,
		Status:  status,
		URL:     fullURL,
	}
}

func (deployment *Deployment) Ready() bool {
	return !deployment.BeingCreated && !deployment.BeingDeleted
}

func CreateDeployment(deploymentID, name, owner string) error {
	currentDeployment, err := GetByID(deploymentID)
	if err != nil {
		return err
	}

	if currentDeployment != nil {
		return nil
	}

	deployment := Deployment{
		ID:           deploymentID,
		Name:         name,
		OwnerID:      owner,
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

func GetByID(deploymentID string) (*Deployment, error) {
	return getDeployment(bson.D{{"id", deploymentID}})
}

func GetByName(name string) (*Deployment, error) {
	return getDeployment(bson.D{{"name", name}})
}

func GetByWebhookToken(token string) (*Deployment, error) {
	return getDeployment(bson.D{{"subsystems.harbor.webhook.token", token}})
}

func Exists(name string) (bool, *Deployment, error) {
	deployment, err := getDeployment(bson.D{{"name", name}})
	if err != nil {
		return false, nil, err
	}

	if deployment == nil {
		return false, nil, nil
	}

	return true, deployment, err
}

func GetMany(ownerID string) ([]Deployment, error) {
	cursor, err := models.DeploymentCollection.Find(context.TODO(), bson.D{{"ownerId", ownerID}})

	if err != nil {
		err = fmt.Errorf("failed to find deployments from owner ID %s. details: %s", ownerID, err)
		log.Println(err)
		return nil, err
	}

	var deployments []Deployment
	for cursor.Next(context.TODO()) {
		var deployment Deployment

		err = cursor.Decode(&deployment)
		if err != nil {
			err = fmt.Errorf("failed to fetch deployment when fetching all deployments from owner ID %s. details: %s", ownerID, err)
			log.Println(err)
			return nil, err
		}
		deployments = append(deployments, deployment)
	}

	return deployments, nil
}

func DeleteByID(deploymentID, userId string) error {
	_, err := models.DeploymentCollection.DeleteOne(context.TODO(), bson.D{{"id", deploymentID}, {"ownerId", userId}})
	if err != nil {
		err = fmt.Errorf("failed to delete deployment %s. details: %s", deploymentID, err)
		log.Println(err)
		return err
	}
	return nil
}

func CountByOwnerID(ownerID string) (int, error) {
	count, err := models.DeploymentCollection.CountDocuments(context.TODO(), bson.D{{"ownerId", ownerID}})

	if err != nil {
		err = fmt.Errorf("failed to count deployments by owner ID %s. details: %s", ownerID, err)
		log.Println(err)
		return 0, err
	}

	return int(count), nil
}

func UpdateByID(id string, update bson.D) error {
	_, err := models.DeploymentCollection.UpdateOne(context.TODO(), bson.D{{"id", id}}, bson.D{{"$set", update}})
	if err != nil {
		err = fmt.Errorf("failed to update deployment %s. details: %s", id, err)
		log.Println(err)
		return err
	}
	return nil
}

func UpdateByName(name string, update bson.D) error {
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
	return UpdateByName(name, bson.D{{subsystemKey, update}})
}

func GetAll() ([]Deployment, error) {
	return GetAllWithFilter(bson.D{})
}

func GetAllWithFilter(filter bson.D) ([]Deployment, error) {
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
