package deployment

import (
	"context"
	"errors"
	"fmt"
	"go-deploy/models"
	"go-deploy/models/sys/activity"
	"go-deploy/models/sys/deployment/subsystems"
	"go-deploy/pkg/app/status_codes"
	"go-deploy/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"sort"
	"time"
)

var (
	NonUniqueFieldErr = fmt.Errorf("non unique field")
)

func (client *Client) Create(id, ownerID string, params *CreateParams) (*Deployment, error) {
	appName := "main"
	replicas := 1
	if params.Replicas != nil {
		replicas = *params.Replicas
	}

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
		Replicas:     replicas,
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
		Activities:  map[string]activity.Activity{ActivityBeingCreated: {ActivityBeingCreated, time.Now()}},
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
		Logs:          make([]Log, 0),
		StatusMessage: status_codes.GetMsg(status_codes.ResourceBeingCreated),
		StatusCode:    status_codes.ResourceBeingCreated,
		Transfer:      nil,
	}

	_, err := client.Collection.InsertOne(context.TODO(), deployment)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, NonUniqueFieldErr
		}

		return nil, fmt.Errorf("failed to create deployment. details: %w", err)
	}

	return client.GetByID(id)
}

func (client *Client) ListByGitHubWebhookID(id int64) ([]Deployment, error) {
	return client.ListWithFilterAndProjection(bson.D{{"subsystems.github.webhook.id", id}}, nil)
}

func (client *Client) GetByTransferCode(code, userID string) (*Deployment, error) {
	return client.GetWithFilterAndProjection(bson.D{{"transfer.code", code}, {"transfer.userId", userID}}, nil)
}

func (client *Client) DeleteByID(id string) error {
	_, err := client.Collection.UpdateOne(context.TODO(),
		bson.D{{"id", id}},
		bson.D{
			{"$set", bson.D{{"deletedAt", time.Now()}}},
			{"$set", bson.D{{"activities", make(map[string]activity.Activity)}}},
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

	setUpdate := bson.D{}
	unsetUpdate := bson.D{}

	models.AddIfNotNil(&setUpdate, "name", params.Name)
	models.AddIfNotNil(&setUpdate, "ownerId", params.OwnerID)
	models.AddIfNotNil(&setUpdate, "apps.main.internalPort", params.InternalPort)
	models.AddIfNotNil(&setUpdate, "apps.main.envs", params.Envs)
	models.AddIfNotNil(&setUpdate, "apps.main.private", params.Private)
	models.AddIfNotNil(&setUpdate, "apps.main.volumes", params.Volumes)
	models.AddIfNotNil(&setUpdate, "apps.main.initCommands", params.InitCommands)
	models.AddIfNotNil(&setUpdate, "apps.main.customDomain", params.CustomDomain)
	models.AddIfNotNil(&setUpdate, "apps.main.image", params.Image)
	models.AddIfNotNil(&setUpdate, "apps.main.pingPath", params.PingPath)
	models.AddIfNotNil(&setUpdate, "apps.main.replicas", params.Replicas)

	if utils.EmptyValue(params.TransferCode) && utils.EmptyValue(params.TransferUserID) {
		models.AddIfNotNil(&unsetUpdate, "transfer", "")
	} else {
		models.AddIfNotNil(&setUpdate, "transfer.code", params.TransferCode)
		models.AddIfNotNil(&setUpdate, "transfer.userId", params.TransferUserID)
	}

	err = client.UpdateWithBsonByID(id,
		bson.D{
			{"$set", setUpdate},
			{"$unset", unsetUpdate},
		},
	)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return NonUniqueFieldErr
		}

		return fmt.Errorf("failed to update deployment %s. details: %w", id, err)
	}

	return nil
}

func (client *Client) GetLogs(id string, history int) ([]Log, error) {
	projection := bson.D{
		{"logs", bson.D{
			{"$slice", -history},
		}},
	}

	var deployment Deployment
	err := client.Collection.FindOne(context.TODO(),
		bson.D{{"id", id}},
		options.FindOne().SetProjection(projection),
	).Decode(&deployment)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return []Log{}, nil
		}

		return nil, err
	}

	return deployment.Logs, nil
}

func (client *Client) GetLogsAfter(id string, createdAt time.Time) ([]Log, error) {
	projection := bson.D{
		{"logs", bson.D{
			{"$slice", -NLogsCache},
		}},
	}

	filter := bson.D{
		{"id", id},
	}

	deployment, err := client.GetWithFilterAndProjection(filter, projection)
	if err != nil {
		return nil, err
	}

	if deployment == nil {
		return nil, nil
	}

	filtered := make([]Log, 0)
	for _, item := range deployment.Logs {
		if item.CreatedAt.After(createdAt) {
			filtered = append(filtered, item)
		}
	}

	return filtered, nil
}

// AddLogs adds logs to the end of the logs array
// Only the last NLogsCache logs are kept
// Logs are sorted by createdAt
func (client *Client) AddLogs(id string, logs ...Log) error {
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].CreatedAt.Before(logs[j].CreatedAt)
	})

	update := bson.D{
		{"$push", bson.D{
			{"logs", bson.D{
				{"$each", logs},
				{"$slice", -NLogsCache},
			}},
		}},
	}

	err := client.UpdateWithBsonByID(id, update)
	if err != nil {
		return err
	}

	return nil
}

func (client *Client) DeleteSubsystemByID(id, key string) error {
	subsystemKey := fmt.Sprintf("subsystems.%s", key)
	return client.UpdateWithBsonByID(id, bson.D{{"$unset", bson.D{{subsystemKey, ""}}}})
}

func (client *Client) UpdateSubsystemByID(id, key string, update interface{}) error {
	subsystemKey := fmt.Sprintf("subsystems.%s", key)
	return client.SetWithBsonByID(id, bson.D{{subsystemKey, update}})
}

func (client *Client) MarkRepaired(id string) error {
	filter := bson.D{{"id", id}}
	update := bson.D{
		{"$set", bson.D{{"repairedAt", time.Now()}}},
		{"$unset", bson.D{{"activities.repairing", ""}}},
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
		{"$unset", bson.D{{"activities.updating", ""}}},
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

func (client *Client) CountReplicas() (int, error) {
	deployments, err := client.List()
	if err != nil {
		return 0, err
	}

	sum := 0
	for _, deployment := range deployments {
		for _, app := range deployment.Apps {
			if app.Replicas > 0 {
				sum += app.Replicas
			} else {
				sum += 1
			}
		}
	}

	return sum, nil
}
