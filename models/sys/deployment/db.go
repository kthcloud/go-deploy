package deployment

import (
	"context"
	"fmt"
	"go-deploy/models/sys/deployment/subsystems"
	status_codes2 "go-deploy/pkg/app/status_codes"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

func (client *Client) Create(id, ownerID string, params *CreateParams) (*Deployment, error) {
	appName := "main"
	mainApp := App{
		Name:         appName,
		Image:        params.Image,
		InternalPort: params.InternalPort,
		Private:      params.Private,
		Envs:         params.Envs,
		Volumes:      params.Volumes,
		InitCommands: params.InitCommands,
		CustomDomain: params.CustomDomain,
		PingResult:   0,
		PingPath:     params.PingPath,
	}

	deployment := Deployment{
		ID:          id,
		Name:        params.Name,
		Type:        params.Type,
		OwnerID:     ownerID,
		Zone:        params.Zone,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Time{},
		RepairedAt:  time.Time{},
		RestartedAt: time.Time{},
		DeletedAt:   time.Time{},
		Activities:  []string{ActivityBeingCreated},
		Apps:        map[string]App{appName: mainApp},
		Subsystems: Subsystems{
			GitLab: subsystems.GitLab{
				LastBuild: subsystems.GitLabBuild{
					ID:        0,
					ProjectID: 0,
					Trace:     []string{"created by go-deploy"},
					Status:    "initialized",
					Stage:     "initialization",
					CreatedAt: time.Now(),
				},
			},
		},
		StatusMessage: status_codes2.GetMsg(status_codes2.ResourceBeingCreated),
		StatusCode:    status_codes2.ResourceBeingCreated,
	}

	filter := bson.D{{"name", params.Name}, {"deletedAt", bson.D{{"$in", []interface{}{time.Time{}, nil}}}}}
	result, err := client.Collection.UpdateOne(context.TODO(), filter, bson.D{
		{"$setOnInsert", deployment},
	}, options.Update().SetUpsert(true))
	if err != nil {
		return nil, fmt.Errorf("failed to create deployment. details: %w", err)
	}

	if result.UpsertedCount == 0 {
		if result.MatchedCount == 1 {
			fetchedDeployment, err := client.GetByName(params.Name)
			if err != nil {
				return nil, err
			}

			if fetchedDeployment == nil {
				log.Println(fmt.Errorf("failed to fetch deployment %s after creation. assuming it was deleted", params.Name))
				return nil, nil
			}

			if fetchedDeployment.ID == id {
				return fetchedDeployment, nil
			}
		}

		return nil, nil
	}

	fetchedDeployment, err := client.GetByName(params.Name)
	if err != nil {
		return nil, err
	}

	return fetchedDeployment, nil
}

func (client *Client) ListByGitHubWebhookID(id int64) ([]Deployment, error) {
	return client.ListWithFilter(bson.D{{"subsystems.github.webhook.id", id}})
}

func (client *Client) DeleteByID(id string) error {
	_, err := client.Collection.UpdateOne(context.TODO(),
		bson.D{{"id", id}},
		bson.D{
			{"$set", bson.D{{"deletedAt", time.Now()}}},
			{"$pull", bson.D{{"activities", ActivityBeingDeleted}}},
		},
	)
	if err != nil {
		return fmt.Errorf("failed to delete deployment %s. details: %w", id, err)
	}

	return nil
}

func (client *Client) UpdateWithParamsByID(id string, params *UpdateParams) error {
	deployment, err := client.GetByID(id)
	if err != nil {
		return err
	}

	if deployment == nil {
		log.Println("deployment not found when updating deployment", id, ". assuming it was deleted")
		return nil
	}

	mainApp := deployment.GetMainApp()
	if mainApp == nil {
		log.Println("main app not found when updating deployment", id, ". assuming it was deleted")
		return nil
	}

	if params.InternalPort != nil {
		mainApp.InternalPort = *params.InternalPort
	}

	if params.Envs != nil {
		mainApp.Envs = *params.Envs
	}

	if params.Private != nil {
		mainApp.Private = *params.Private
	}

	if params.CustomDomain != nil {
		mainApp.CustomDomain = params.CustomDomain
	}

	if params.Volumes != nil {
		mainApp.Volumes = *params.Volumes
	}

	if params.InitCommands != nil {
		mainApp.InitCommands = *params.InitCommands
	}

	if params.Image != nil {
		mainApp.Image = *params.Image
	}

	if params.PingPath != nil {
		mainApp.PingPath = *params.PingPath
	}

	deployment.Apps["main"] = *mainApp

	_, err = client.Collection.UpdateOne(context.TODO(),
		bson.D{{"id", id}},
		bson.D{{"$set", bson.D{{"apps", deployment.Apps}}}},
	)
	if err != nil {
		return fmt.Errorf("failed to update deployment %s. details: %w", id, err)
	}

	return nil

}

func (client *Client) DeleteSubsystemByID(id, key string) error {
	subsystemKey := fmt.Sprintf("subsystems.%s", key)
	return client.UpdateWithBsonByID(id, bson.D{{"$unset", bson.D{{subsystemKey, ""}}}})
}

func (client *Client) DeleteSubsystemByName(name, key string) error {
	subsystemKey := fmt.Sprintf("subsystems.%s", key)
	return client.UpdateWithBsonByName(name, bson.D{{"$unset", bson.D{{subsystemKey, ""}}}})
}

func (client *Client) UpdateSubsystemByName(name, key string, update interface{}) error {
	subsystemKey := fmt.Sprintf("subsystems.%s", key)
	return client.SetWithBsonByName(name, bson.D{{subsystemKey, update}})
}

func (client *Client) UpdateSubsystemByID(id, key string, update interface{}) error {
	subsystemKey := fmt.Sprintf("subsystems.%s", key)
	return client.SetWithBsonByID(id, bson.D{{subsystemKey, update}})
}

func (client *Client) MarkRepaired(id string) error {
	filter := bson.D{{"id", id}}
	update := bson.D{
		{"$set", bson.D{{"repairedAt", time.Now()}}},
		{"$pull", bson.D{{"activities", "repairing"}}},
	}

	_, err := client.Collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (client *Client) MarkUpdated(id string) error {
	filter := bson.D{{"id", id}}
	update := bson.D{
		{"$set", bson.D{{"updatedAt", time.Now()}}},
		{"$pull", bson.D{{"activities", "updating"}}},
	}

	_, err := client.Collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (client *Client) UpdateGitLabBuild(deploymentID string, build subsystems.GitLabBuild) error {
	filter := bson.D{
		{"id", deploymentID},
		{"subsystems.gitlab.lastBuild.createdAt", bson.M{"$lte": build.CreatedAt}},
	}

	update := bson.D{
		{"$set", bson.D{
			{"subsystems.gitlab.lastBuild", build},
		}},
	}

	_, err := client.Collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

func (client *Client) GetLastGitLabBuild(deploymentID string) (*subsystems.GitLabBuild, error) {
	// fetch only subsystem.gitlab.lastBuild
	projection := bson.D{
		{"subsystems.gitlab.lastBuild", 1},
	}

	var deployment Deployment
	err := client.Collection.FindOne(context.TODO(),
		bson.D{{"id", deploymentID}},
		options.FindOne().SetProjection(projection),
	).Decode(&deployment)
	if err != nil {
		return &subsystems.GitLabBuild{}, err
	}

	return &deployment.Subsystems.GitLab.LastBuild, nil
}

func (client *Client) SavePing(id string, pingResult int) error {
	deployment, err := client.GetByID(id)
	if err != nil {
		return err
	}

	if deployment == nil {
		log.Println("deployment not found when saving ping result", id, ". assuming it was deleted")
		return nil
	}

	app := deployment.GetMainApp()
	if app == nil {
		return fmt.Errorf("failed to find main app for deployment %s", id)
	}

	app.PingResult = pingResult

	deployment.Apps["main"] = *app

	_, err = client.Collection.UpdateOne(context.TODO(),
		bson.D{{"id", id}},
		bson.D{{"$set", bson.D{{"apps.main.pingResult", pingResult}}}},
	)
	if err != nil {
		return fmt.Errorf("failed to update deployment ping result %s. details: %w", id, err)
	}

	return nil
}

func (client *Client) RemoveCustomDomain(deploymentID string) error {
	return client.SetWithBsonByID(deploymentID, bson.D{{"apps.main.customDomain", nil}})
}
