package vm

import (
	"context"
	"fmt"
	"go-deploy/models"
	"go-deploy/models/sys/activity"
	"go-deploy/pkg/app/status_codes"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

var (
	NonUniqueFieldErr = fmt.Errorf("non unique field")
)

func (client *Client) Create(id, owner, manager string, params *CreateParams) (*VM, error) {
	var ports []Port
	if params.Ports != nil {
		ports = params.Ports
	} else {
		ports = make([]Port, 0)
	}

	vm := VM{
		ID:        id,
		Name:      params.Name,
		OwnerID:   owner,
		ManagedBy: manager,

		Zone:           params.Zone,
		DeploymentZone: params.DeploymentZone,

		CreatedAt:  time.Now(),
		UpdatedAt:  time.Time{},
		RepairedAt: time.Time{},
		DeletedAt:  time.Time{},

		GpuID:        "",
		SshPublicKey: params.SshPublicKey,
		Ports:        ports,
		Activities:   map[string]activity.Activity{ActivityBeingCreated: {ActivityBeingCreated, time.Now()}},
		Subsystems:   Subsystems{},
		Specs: Specs{
			CpuCores: params.CpuCores,
			RAM:      params.RAM,
			DiskSize: params.DiskSize,
		},

		StatusCode:    status_codes.ResourceBeingCreated,
		StatusMessage: status_codes.GetMsg(status_codes.ResourceBeingCreated),
	}

	_, err := client.Collection.InsertOne(context.TODO(), vm)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, NonUniqueFieldErr
		}

		return nil, fmt.Errorf("failed to create vm %s. details: %w", id, err)
	}

	return client.GetByID(id)
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

func (client *Client) UpdateWithParamsByID(id string, update *UpdateParams) error {
	updateData := bson.D{}

	models.AddIfNotNil(&updateData, "name", update.Name)
	models.AddIfNotNil(&updateData, "ports", update.Ports)
	models.AddIfNotNil(&updateData, "specs.cpuCores", update.CpuCores)
	models.AddIfNotNil(&updateData, "specs.ram", update.RAM)

	if len(updateData) == 0 {
		return nil
	}

	_, err := client.Collection.UpdateOne(context.TODO(),
		bson.D{{"id", id}},
		bson.D{{"$set", updateData}},
	)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return NonUniqueFieldErr
		}

		return fmt.Errorf("failed to update vm %s. details: %w", id, err)
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

func (client *Client) ListWithGPU() ([]VM, error) {
	// create a filter that checks if the gpuID field is not empty
	filter := bson.D{{
		"gpuId", bson.M{
			"$ne": "",
		},
	}}

	return client.ListWithFilter(filter)
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
