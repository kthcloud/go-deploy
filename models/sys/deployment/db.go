package deployment

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
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

// Create creates a new deployment with the given params.
// It will return a NonUniqueFieldErr any unique index was violated.
func (client *Client) Create(id, ownerID string, params *CreateParams) (*Deployment, error) {
	appName := "main"
	replicas := 1
	if params.Replicas != nil {
		replicas = *params.Replicas
	}

	var customDomain *CustomDomain
	if params.CustomDomain != nil {
		customDomain = &CustomDomain{
			Domain: *params.CustomDomain,
			Secret: generateCustomDomainSecret(),
			Status: CustomDomainStatusPending,
		}
	}

	mainApp := App{
		Name:         appName,
		Image:        params.Image,
		InternalPort: params.InternalPort,
		Private:      params.Private,
		Envs:         params.Envs,
		Volumes:      params.Volumes,
		InitCommands: params.InitCommands,
		CustomDomain: customDomain,
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

// UpdateWithParams updates a deployment with the given update params.
// It will return a NonUniqueFieldErr any unique index was violated.
func (client *Client) UpdateWithParams(id string, params *UpdateParams) error {
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

	// If the custom domain is empty, it means we want to remove it
	var customDomain *CustomDomain
	if params.CustomDomain != nil {
		if *params.CustomDomain == "" {
			models.Add(&unsetUpdate, fmt.Sprintf("apps.%s.customDomain", mainApp.Name), "")
		} else {
			models.AddIfNotNil(&setUpdate, fmt.Sprintf("apps.%s.customDomain", mainApp.Name), &CustomDomain{
				Domain: *params.CustomDomain,
				Secret: generateCustomDomainSecret(),
				Status: CustomDomainStatusPending,
			})
		}
	}

	// If the transfer code is empty, it means it is either done and we remove it,
	// or we don't want to transfer it anymore
	if utils.EmptyValue(params.TransferCode) && utils.EmptyValue(params.TransferUserID) {
		models.AddIfNotNil(&unsetUpdate, "transfer", "")
	} else {
		models.AddIfNotNil(&setUpdate, "transfer.code", params.TransferCode)
		models.AddIfNotNil(&setUpdate, "transfer.userId", params.TransferUserID)
	}

	models.AddIfNotNil(&setUpdate, "name", params.Name)
	models.AddIfNotNil(&setUpdate, "ownerId", params.OwnerID)
	models.AddIfNotNil(&setUpdate, "apps.main.internalPort", params.InternalPort)
	models.AddIfNotNil(&setUpdate, "apps.main.envs", params.Envs)
	models.AddIfNotNil(&setUpdate, "apps.main.private", params.Private)
	models.AddIfNotNil(&setUpdate, "apps.main.volumes", params.Volumes)
	models.AddIfNotNil(&setUpdate, "apps.main.initCommands", params.InitCommands)
	models.AddIfNotNil(&setUpdate, "apps.main.customDomain", customDomain)
	models.AddIfNotNil(&setUpdate, "apps.main.image", params.Image)
	models.AddIfNotNil(&setUpdate, "apps.main.pingPath", params.PingPath)
	models.AddIfNotNil(&setUpdate, "apps.main.replicas", params.Replicas)

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

// GetLogs returns the last NLogsCache logs for a deployment
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

// GetLogsAfter returns the logs after the given time, with a maximum of NLogsCache logs.
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

// AddLogs adds logs to the end of the log array.
// Only the last NLogsCache logs are kept, and are sorted by createdAt
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

// DeleteSubsystem erases a subsystem from a deployment.
// It prepends the key with `subsystems` and unsets it.
func (client *Client) DeleteSubsystem(id, key string) error {
	subsystemKey := fmt.Sprintf("subsystems.%s", key)
	return client.UpdateWithBsonByID(id, bson.D{{"$unset", bson.D{{subsystemKey, ""}}}})
}

// SetSubsystem updates a subsystem from a deployment.
// It prepends the key with `subsystems` and sets it.
func (client *Client) SetSubsystem(id, key string, update interface{}) error {
	subsystemKey := fmt.Sprintf("subsystems.%s", key)
	return client.SetWithBsonByID(id, bson.D{{subsystemKey, update}})
}

// MarkRepaired marks a deployment as repaired.
// It sets RepairedAt and unsets the repairing activity.
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

// MarkUpdated marks a deployment as updated.
// It sets UpdatedAt and unsets the updating activity.
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

// UpdateCustomDomainStatus updates the status of a custom domain, such as
// CustomDomainStatusActive, CustomDomainStatusVerificationFailed, CustomDomainStatusPending
func (client *Client) UpdateCustomDomainStatus(id string, status string) error {
	filter := bson.D{{"id", id}}
	update := bson.D{
		{"$set", bson.D{{"apps.main.customDomain.status", status}}},
	}

	_, err := client.Collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

// UpdateGitLabBuild updates the last GitLab build for a deployment.
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

// GetLastGitLabBuild returns the last GitLab build for a deployment.
func (client *Client) GetLastGitLabBuild(deploymentID string) (*subsystems.GitLabBuild, error) {
	deployment, err := client.GetWithFilterAndProjection(bson.D{{"id", deploymentID}}, bson.D{{"subsystems.gitlab.lastBuild", 1}})
	if err != nil {
		return nil, err
	}

	if deployment == nil {
		return nil, nil
	}

	return &deployment.Subsystems.GitLab.LastBuild, nil
}

// SetPingResult sets the ping result for a deployment.
func (client *Client) SetPingResult(id string, pingResult int) error {
	exists, err := client.ExistsByID(id)
	if err != nil {
		return err
	}

	if !exists {
		log.Println("deployment not found when saving ping result", id, ". assuming it was deleted")
		return nil
	}

	err = client.UpdateWithBsonByID(id, bson.D{{"$set", bson.D{{"apps.main.pingResult", pingResult}}}})
	if err != nil {
		return fmt.Errorf("failed to update deployment ping result %s. details: %w", id, err)
	}

	return nil
}

// CountReplicas returns the number of replicas for all apps in all deployments.
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

// generateCustomDomainSecret generates a random alphanumeric string.
func generateCustomDomainSecret() string {
	return utils.HashStringAlphanumeric(uuid.NewString())
}
