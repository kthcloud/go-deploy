package vm

import (
	"context"
	"fmt"
	"go-deploy/models"
	"go-deploy/models/sys/activity"
	"go-deploy/pkg/app/status_codes"
	"go-deploy/utils"
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

func (client *Client) GetByTransferCode(code, userID string) (*VM, error) {
	return client.GetWithFilter(bson.D{{"transfer.code", code}, {"transfer.userId", userID}})
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
	setUpdate := bson.D{}
	unsetUpdate := bson.D{}

	// check that no other vm has a port with the same http proxy name
	// this should ideally be done with unique indexes, but the indexes are per document, and not per-array element.
	// so we would need to create a separate collection for ports, which is not ideal, since that would require
	// us to use $lookup to get the ports for a vm or use transactions, which is quite annoying :)
	// so we do it manually here, even though it might cause a race condition, because I'm lazy.
	if params.Ports != nil {
		for _, port := range *params.Ports {
			if port.HttpProxy != nil {
				filter := bson.D{
					{"id", bson.D{{"$ne", id}}},
					{"ports.httpProxy.name", port.HttpProxy.Name},
				}

				existAny, err := client.ResourceClient.AddExtraFilter(filter).ExistsAny()
				if err != nil {
					return err
				}

				if existAny {
					return NonUniqueFieldErr
				}
			}
		}
	}

	models.AddIfNotNil(&setUpdate, "name", params.Name)
	models.AddIfNotNil(&setUpdate, "ownerId", params.OwnerID)
	models.AddIfNotNil(&setUpdate, "ports", params.Ports)
	models.AddIfNotNil(&setUpdate, "specs.cpuCores", params.CpuCores)
	models.AddIfNotNil(&setUpdate, "specs.ram", params.RAM)

	if utils.EmptyValue(params.TransferCode) && utils.EmptyValue(params.TransferUserID) {
		models.AddIfNotNil(&unsetUpdate, "transfer", "")
	} else {
		models.AddIfNotNil(&setUpdate, "transfer.code", params.TransferCode)
		models.AddIfNotNil(&setUpdate, "transfer.userId", params.TransferUserID)
	}

	err := client.UpdateWithBsonByID(id,
		bson.D{
			{"$set", setUpdate},
			{"$unset", unsetUpdate},
		},
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
