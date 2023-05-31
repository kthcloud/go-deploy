package deployment

import (
	"context"
	"fmt"
	"go-deploy/models"
	"go-deploy/pkg/status_codes"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"time"
)

func (deployment *Deployment) Ready() bool {
	return !deployment.DoingActivity(ActivityBeingCreated) && !deployment.DoingActivity(ActivityBeingDeleted)
}

func (deployment *Deployment) DoingActivity(activity string) bool {
	for _, a := range deployment.Activities {
		if a == activity {
			return true
		}
	}
	return false
}

func (deployment *Deployment) BeingCreated() bool {
	return deployment.DoingActivity(ActivityBeingCreated)
}

func (deployment *Deployment) BeingDeleted() bool {
	return deployment.DoingActivity(ActivityBeingDeleted)
}

func CreateDeployment(deploymentID, ownerID string, params *CreateParams) error {
	currentDeployment, err := GetByID(deploymentID)
	if err != nil {
		return err
	}

	if currentDeployment != nil {
		return nil
	}

	deployment := Deployment{
		ID:      deploymentID,
		Name:    params.Name,
		OwnerID: ownerID,

		Private:      params.Private,
		Envs:         params.Envs,
		ExtraDomains: make([]string, 0),

		Activities: []string{ActivityBeingCreated},
		RepairedAt: time.Now(),

		StatusCode:    status_codes.ResourceBeingCreated,
		StatusMessage: status_codes.GetMsg(status_codes.ResourceBeingCreated),
	}

	_, err = models.DeploymentCollection.InsertOne(context.TODO(), deployment)
	if err != nil {
		err = fmt.Errorf("failed to create deployment %s. details: %s", params.Name, err)
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

func GetByGitHubWebhookID(id int64) (*Deployment, error) {
	return getDeployment(bson.D{{"subsystems.github.webhook.id", id}})
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
		return nil, err
	}

	var deployments []Deployment
	for cursor.Next(context.TODO()) {
		var deployment Deployment

		err = cursor.Decode(&deployment)
		if err != nil {
			err = fmt.Errorf("failed to fetch deployment when fetching all deployments from owner ID %s. details: %s", ownerID, err)
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
		return 0, err
	}

	return int(count), nil
}

func UpdateByID(id string, update *UpdateParams) error {
	updateData := bson.M{}

	models.AddIfNotNil(updateData, "envs", update.Envs)
	models.AddIfNotNil(updateData, "private", update.Private)
	models.AddIfNotNil(updateData, "extraDomains", update.ExtraDomains)

	if len(updateData) == 0 {
		return nil
	}

	_, err := models.DeploymentCollection.UpdateOne(context.TODO(),
		bson.D{{"id", id}},
		bson.D{{"$set", updateData}},
	)
	if err != nil {
		return fmt.Errorf("failed to update deployment %s. details: %s", id, err)
	}

	return nil

}

func UpdateWithBsonByID(id string, update bson.D) error {
	_, err := models.DeploymentCollection.UpdateOne(context.TODO(), bson.D{{"id", id}}, bson.D{{"$set", update}})
	if err != nil {
		err = fmt.Errorf("failed to update deployment %s. details: %s", id, err)
		return err
	}
	return nil
}

func UpdateByName(name string, update bson.D) error {
	_, err := models.DeploymentCollection.UpdateOne(context.TODO(), bson.D{{"name", name}}, bson.D{{"$set", update}})
	if err != nil {
		err = fmt.Errorf("failed to update deployment %s. details: %s", name, err)
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
			return nil, err
		}
		deployments = append(deployments, deployment)
	}

	return deployments, nil
}

func GetByActivity(activity string) ([]Deployment, error) {
	filter := bson.D{
		{
			"activities", bson.M{
				"$in": bson.A{activity},
			},
		},
	}

	return GetAllWithFilter(filter)
}

func GetWithNoActivities() ([]Deployment, error) {
	filter := bson.D{
		{
			"activities", bson.M{
				"$size": 0,
			},
		},
	}

	return GetAllWithFilter(filter)
}

func AddActivity(deploymentID, activity string) error {
	_, err := models.DeploymentCollection.UpdateOne(context.TODO(),
		bson.D{{"id", deploymentID}},
		bson.D{{"$addToSet", bson.D{{"activities", activity}}}},
	)
	if err != nil {
		err = fmt.Errorf("failed to add activity %s to deployment %s. details: %s", activity, deploymentID, err)
		return err
	}
	return nil
}

func RemoveActivity(deploymentID, activity string) error {
	_, err := models.DeploymentCollection.UpdateOne(context.TODO(),
		bson.D{{"id", deploymentID}},
		bson.D{{"$pull", bson.D{{"activities", activity}}}},
	)
	if err != nil {
		err = fmt.Errorf("failed to remove activity %s from deployment %s. details: %s", activity, deploymentID, err)
		return err
	}
	return nil
}

func (deployment *Deployment) Created() bool {
	return deployment.ID != "" &&
		deployment.Subsystems.GitHub.Created() &&
		deployment.Subsystems.Harbor.Created() &&
		deployment.Subsystems.K8s.Created()
}

func (k8s *K8s) Created() bool {
	return k8s.Namespace.Created() &&
		k8s.Deployment.Created() &&
		k8s.Service.Created() &&
		k8s.Ingress.Created()
}

func (harbor *Harbor) Created() bool {
	return harbor.Project.Created() &&
		harbor.Repository.Created() &&
		harbor.Robot.Created() &&
		harbor.Webhook.Created()
}

func (gitHub *GitHub) Created() bool {
	return gitHub.Webhook.Created()
}
