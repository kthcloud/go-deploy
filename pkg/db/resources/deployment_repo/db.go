package deployment_repo

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/google/uuid"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/app/status_codes"
	"github.com/kthcloud/go-deploy/pkg/db"
	"github.com/kthcloud/go-deploy/pkg/log"
	rErrors "github.com/kthcloud/go-deploy/service/errors"
	"github.com/kthcloud/go-deploy/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Create creates a new deployment with the given params.
// It will return a ErrNonUniqueField any unique index was violated.
func (client *Client) Create(id, ownerID string, params *model.DeploymentCreateParams) (*model.Deployment, error) {
	appName := "main"

	var customDomain *model.CustomDomain
	if params.CustomDomain != nil {
		customDomain = &model.CustomDomain{
			Domain: *params.CustomDomain,
			Secret: generateCustomDomainSecret(),
			Status: model.CustomDomainStatusPending,
		}
	}

	mainApp := model.App{
		Name: appName,

		CpuCores: params.CpuCores,
		RAM:      params.RAM,
		Replicas: params.Replicas,

		Image:        params.Image,
		InternalPort: params.InternalPort,
		Envs:         params.Envs,
		Volumes:      params.Volumes,
		Visibility:   params.Visibility,

		Args:         params.Args,
		InitCommands: params.InitCommands,
		CustomDomain: customDomain,

		ReplicaStatus: nil,
		PingPath:      params.PingPath,
		PingResult:    0,
	}

	deployment := model.Deployment{
		ID:      id,
		Name:    params.Name,
		Type:    params.Type,
		OwnerID: ownerID,
		Zone:    params.Zone,

		CreatedAt:   time.Now(),
		UpdatedAt:   time.Time{},
		RepairedAt:  time.Time{},
		RestartedAt: time.Time{},
		DeletedAt:   time.Time{},
		AccessedAt:  time.Now(),

		Activities: map[string]model.Activity{model.ActivityBeingCreated: {
			Name:      model.ActivityBeingCreated,
			CreatedAt: time.Now(),
		}},
		Apps:       map[string]model.App{appName: mainApp},
		Subsystems: model.DeploymentSubsystems{},
		Logs:       make([]model.Log, 0),
		Status:     status_codes.GetMsg(status_codes.ResourceCreating),
	}

	_, err := client.Collection.InsertOne(context.TODO(), deployment)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, rErrors.ErrNonUniqueField
		}

		return nil, fmt.Errorf("failed to create deployment. details: %w", err)
	}

	return client.GetByID(id)
}

// UpdateWithParams updates a deployment with the given update params.
// It will return a ErrNonUniqueField any unique index was violated.
func (client *Client) UpdateWithParams(id string, params *model.DeploymentUpdateParams) error {
	deployment, err := client.GetByID(id)
	if err != nil {
		return err
	}

	if deployment == nil {
		log.Println("Deployment not found when updating deployment", id, ". Assuming it was deleted")
		return nil
	}

	mainApp := deployment.GetMainApp()
	if mainApp == nil {
		log.Println("Main app not found when updating deployment", id, ". Assuming it was deleted")
		return nil
	}

	setUpdate := bson.D{}
	unsetUpdate := bson.D{}

	var customDomain *model.CustomDomain
	if params.CustomDomain != nil {
		if *params.CustomDomain == "" {
			// If the custom domain is empty, it means we want to remove it
			db.Add(&unsetUpdate, fmt.Sprintf("apps.%s.customDomain", mainApp.Name), "")
		} else {
			db.AddIfNotNil(&setUpdate, fmt.Sprintf("apps.%s.customDomain", mainApp.Name), &model.CustomDomain{
				Domain: *params.CustomDomain,
				Secret: generateCustomDomainSecret(),
				Status: model.CustomDomainStatusPending,
			})
		}
	}

	db.AddIfNotNil(&setUpdate, "name", params.Name)
	db.AddIfNotNil(&setUpdate, "ownerId", params.OwnerID)
	db.AddIfNotNil(&setUpdate, "apps.main.internalPort", params.InternalPort)
	db.AddIfNotNil(&setUpdate, "apps.main.envs", params.Envs)
	db.AddIfNotNil(&setUpdate, "apps.main.volumes", params.Volumes)
	db.AddIfNotNil(&setUpdate, "apps.main.initCommands", params.InitCommands)
	db.AddIfNotNil(&setUpdate, "apps.main.args", params.Args)
	db.AddIfNotNil(&setUpdate, "apps.main.customDomain", customDomain)
	db.AddIfNotNil(&setUpdate, "apps.main.image", params.Image)
	db.AddIfNotNil(&setUpdate, "apps.main.pingPath", params.PingPath)
	db.AddIfNotNil(&setUpdate, "apps.main.cpuCores", params.CpuCores)
	db.AddIfNotNil(&setUpdate, "apps.main.ram", params.RAM)
	db.AddIfNotNil(&setUpdate, "apps.main.replicas", params.Replicas)
	db.AddIfNotNil(&setUpdate, "apps.main.visibility", params.Visibility)

	err = client.UpdateWithBsonByID(id,
		bson.D{
			{Key: "$set", Value: setUpdate},
			{Key: "$unset", Value: unsetUpdate},
		},
	)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return rErrors.ErrNonUniqueField
		}

		return fmt.Errorf("failed to update deployment %s. details: %w", id, err)
	}

	return nil
}

// GetLogs returns the last NLogsCache logs for a deployment
func (client *Client) GetLogs(id string, history int) ([]model.Log, error) {
	projection := bson.D{
		{Key: "logs", Value: bson.D{
			{Key: "$slice", Value: -history},
		}},
	}

	var deployment model.Deployment
	err := client.Collection.FindOne(context.TODO(),
		bson.D{{Key: "id", Value: id}},
		options.FindOne().SetProjection(projection),
	).Decode(&deployment)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return []model.Log{}, nil
		}

		return nil, err
	}

	return deployment.Logs, nil
}

// GetLogsAfter returns the logs after the given time, with a maximum of NLogsCache logs.
func (client *Client) GetLogsAfter(id string, createdAt time.Time) ([]model.Log, error) {
	projection := bson.D{
		{Key: "logs", Value: bson.D{
			{Key: "$slice", Value: -model.NLogsCache},
		}},
	}

	filter := bson.D{
		{Key: "id", Value: id},
	}

	deployment, err := client.GetWithFilterAndProjection(filter, projection)
	if err != nil {
		return nil, err
	}

	if deployment == nil {
		return nil, nil
	}

	filtered := make([]model.Log, 0)
	for _, item := range deployment.Logs {
		if item.CreatedAt.After(createdAt) {
			filtered = append(filtered, item)
		}
	}

	return filtered, nil
}

// AddLogsByName adds logs to the end of the log array.
// Only the last NLogsCache logs are kept, and are sorted by createdAt
func (client *Client) AddLogsByName(name string, logs ...model.Log) error {
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].CreatedAt.Before(logs[j].CreatedAt)
	})

	update := bson.D{
		{Key: "$push", Value: bson.D{
			{Key: "logs", Value: bson.D{
				{Key: "$each", Value: logs},
				{Key: "$slice", Value: -model.NLogsCache},
			}},
		}},
	}

	err := client.UpdateWithBsonByName(name, update)
	if err != nil {
		return err
	}

	return nil
}

// GetUsage returns the total usage of all deployments.
func (client *Client) GetUsage() (*model.DeploymentUsage, error) {
	projection := bson.D{
		{Key: "_id", Value: 0},
		{Key: "apps.main.replicas", Value: 1},
		{Key: "apps.main.cpuCores", Value: 1},
		{Key: "apps.main.ram", Value: 1},
	}

	deployments, err := client.ListWithFilterAndProjection(bson.D{}, projection)
	if err != nil {
		return nil, err
	}

	usage := &model.DeploymentUsage{}

	for _, deployment := range deployments {
		for _, app := range deployment.Apps {
			usage.CpuCores += app.CpuCores * float64(app.Replicas)
			usage.RAM += app.RAM * float64(app.Replicas)
		}
	}

	return usage, nil
}

// SetStatusByName sets the status of a deployment.
func (client *Client) SetStatusByName(name, status string, replicaStatus *model.ReplicaStatus) error {
	update := bson.D{
		{Key: "status", Value: status},
		{Key: "apps.main.replicaStatus", Value: replicaStatus},
	}

	return client.SetWithBsonByName(name, update)
}

// SetErrorByName sets the error of a deployment.
func (client *Client) SetErrorByName(name string, error *model.DeploymentError) error {
	return client.SetWithBsonByName(name, bson.D{{Key: "error", Value: error}})
}

// UnsetErrorByName unsets the error of a deployment.
func (client *Client) UnsetErrorByName(name string) error {
	return client.UnsetByName(name, "error")
}

// DeleteSubsystem erases a subsystem from a deployment.
// It prepends the key with `subsystems` and unsets it.
func (client *Client) DeleteSubsystem(id, key string) error {
	subsystemKey := fmt.Sprintf("subsystems.%s", key)
	return client.UpdateWithBsonByID(id, bson.D{{Key: "$unset", Value: bson.D{{Key: subsystemKey, Value: ""}}}})
}

// SetSubsystem updates a subsystem from a deployment.
// It prepends the key with `subsystems` and sets it.
func (client *Client) SetSubsystem(id, key string, update interface{}) error {
	subsystemKey := fmt.Sprintf("subsystems.%s", key)
	return client.SetWithBsonByID(id, bson.D{{Key: subsystemKey, Value: update}})
}

// MarkRepaired marks a deployment as repaired.
// It sets RepairedAt and unsets the repairing activity.
func (client *Client) MarkRepaired(id string) error {
	filter := bson.D{{Key: "id", Value: id}}
	update := bson.D{
		{Key: "$set", Value: bson.D{{Key: "repairedAt", Value: time.Now()}}},
		{Key: "$unset", Value: bson.D{{Key: "activities.repairing", Value: ""}}},
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
	filter := bson.D{{Key: "id", Value: id}}
	update := bson.D{
		{Key: "$set", Value: bson.D{{Key: "updatedAt", Value: time.Now()}}},
		{Key: "$unset", Value: bson.D{{Key: "activities.updating", Value: ""}}},
	}

	_, err := client.Collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

// MarkAccessed marks a deployment as accessed to the current time.
func (client *Client) MarkAccessed(id string) error {
	return client.SetWithBsonByID(id, bson.D{{Key: "accessedAt", Value: time.Now()}})
}

// UpdateCustomDomainStatus updates the status of a custom domain, such as
// CustomDomainStatusActive, CustomDomainStatusVerificationFailed, CustomDomainStatusPending
func (client *Client) UpdateCustomDomainStatus(id string, status string) error {
	filter := bson.D{{Key: "id", Value: id}}
	update := bson.D{
		{Key: "$set", Value: bson.D{{Key: "apps.main.customDomain.status", Value: status}}},
	}

	_, err := client.Collection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

// SetPingResult sets the ping result for a deployment.
func (client *Client) SetPingResult(id string, pingResult int) error {
	exists, err := client.ExistsByID(id)
	if err != nil {
		return err
	}

	if !exists {
		log.Println("Deployment not found when saving ping result", id, ". Assuming it was deleted")
		return nil
	}

	err = client.UpdateWithBsonByID(id, bson.D{{Key: "$set", Value: bson.D{{Key: "apps.main.pingResult", Value: pingResult}}}})
	if err != nil {
		return fmt.Errorf("failed to update deployment ping result %s. details: %w", id, err)
	}

	return nil
}

// generateCustomDomainSecret generates a random alphanumeric string.
func generateCustomDomainSecret() string {
	return utils.HashStringAlphanumeric(uuid.NewString())
}
