package deployment

import (
	"context"
	"fmt"
	"go-deploy/models"
	"go-deploy/pkg/status_codes"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

func CreateDeployment(deploymentID, ownerID string, params *CreateParams) (bool, error) {
	deployment := Deployment{
		ID:      deploymentID,
		Name:    params.Name,
		OwnerID: ownerID,

		CreatedAt: time.Now(),

		Private:      params.Private,
		Envs:         params.Envs,
		ExtraDomains: make([]string, 0),

		Activities: []string{ActivityBeingCreated},

		Subsystems: Subsystems{
			GitLab: GitLab{
				LastBuild: GitLabBuild{
					ID:        0,
					ProjectID: 0,
					Trace:     []string{"created by go-deploy"},
					Status:    "initialized",
					Stage:     "initialization",
					CreatedAt: time.Now(),
				},
			},
		},

		StatusCode:    status_codes.ResourceBeingCreated,
		StatusMessage: status_codes.GetMsg(status_codes.ResourceBeingCreated),

		ZoneID: params.ZoneID,
	}

	result, err := models.DeploymentCollection.UpdateOne(context.TODO(), bson.D{{"name", params.Name}}, bson.D{
		{"$setOnInsert", deployment},
	}, options.Update().SetUpsert(true))
	if err != nil {
		return false, fmt.Errorf("failed to create deployment. details: %s", err)
	}

	if result.UpsertedCount == 0 {
		if result.MatchedCount == 1 {
			fetchedDeployment, err := getDeployment(bson.D{{"name", params.Name}})
			if err != nil {
				return false, err
			}

			if fetchedDeployment.ID == deploymentID {
				return true, nil
			}
		}

		return false, nil
	}

	return true, nil
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

func GetAllByGitHubWebhookID(id int64) ([]Deployment, error) {
	filter := bson.D{{"subsystems.github.webhook.id", id}}

	cursor, err := models.DeploymentCollection.Find(context.TODO(), filter)
	if err != nil {
		err = fmt.Errorf("failed to find deployments by GitHub webhook ID %d. details: %s", id, err)
		return nil, err
	}

	var deployments []Deployment
	for cursor.Next(context.TODO()) {
		var deployment Deployment

		err = cursor.Decode(&deployment)
		if err != nil {
			err = fmt.Errorf("failed to fetch deployment. details: %s", err)
			return nil, err
		}

		deployments = append(deployments, deployment)
	}

	return deployments, nil
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

func ClearActivities(deploymentID string) error {
	_, err := models.DeploymentCollection.UpdateOne(context.TODO(),
		bson.D{{"id", deploymentID}},
		bson.D{{"$set", bson.D{{"activities", bson.A{}}}}},
	)
	if err != nil {
		err = fmt.Errorf("failed to clear activities from deployment %s. details: %s", deploymentID, err)
		return err
	}

	return nil
}

func MarkRepaired(deploymentID string) error {
	filter := bson.D{{"id", deploymentID}}
	update := bson.D{
		{"$set", bson.D{{"repairedAt", time.Now()}}},
		{"$pull", bson.D{{"activities", "repairing"}}},
	}

	_, err := models.DeploymentCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

func MarkUpdated(deploymentID string) error {
	filter := bson.D{{"id", deploymentID}}
	update := bson.D{
		{"$set", bson.D{{"updatedAt", time.Now()}}},
		{"$pull", bson.D{{"activities", "updating"}}},
	}

	_, err := models.DeploymentCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

func UpdateGitLabBuild(deploymentID string, build GitLabBuild) error {
	filter := bson.D{
		{"id", deploymentID},
		{"subsystems.gitlab.lastBuild.createdAt", bson.M{"$lte": build.CreatedAt}},
	}

	update := bson.D{
		{"$set", bson.D{
			{"subsystems.gitlab.lastBuild", build},
		}},
	}

	_, err := models.DeploymentCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

func GetLastGitLabBuild(deploymentID string) (*GitLabBuild, error) {
	// fetch only subsystem.gitlab.lastBuild
	projection := bson.D{
		{"subsystems.gitlab.lastBuild", 1},
	}

	var deployment Deployment
	err := models.DeploymentCollection.FindOne(context.TODO(),
		bson.D{{"id", deploymentID}},
		options.FindOne().SetProjection(projection),
	).Decode(&deployment)
	if err != nil {
		return &GitLabBuild{}, err
	}

	return &deployment.Subsystems.GitLab.LastBuild, nil
}

func SavePing(id string, pingResult int) error {
	updateData := bson.M{}

	models.AddIfNotNil(updateData, "pingResult", pingResult)

	if len(updateData) == 0 {
		return nil
	}

	_, err := models.DeploymentCollection.UpdateOne(context.TODO(),
		bson.D{{"id", id}},
		bson.D{{"$set", updateData}},
	)
	if err != nil {
		return fmt.Errorf("failed to update deployment ping result %s. details: %s", id, err)
	}

	return nil

}
