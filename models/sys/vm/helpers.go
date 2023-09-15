package vm

import (
	"context"
	"fmt"
	"go-deploy/models"
	"go-deploy/pkg/status_codes"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"time"
)

func (client *Client) Create(vmID, owner, manager string, params *CreateParams) (bool, error) {
	var ports []Port
	if params.Ports != nil {
		ports = params.Ports
	} else {
		ports = make([]Port, 0)
	}

	vm := VM{
		ID:        vmID,
		Name:      params.Name,
		OwnerID:   owner,
		ManagedBy: manager,
		Zone:      params.Zone,

		CreatedAt:  time.Now(),
		UpdatedAt:  time.Time{},
		RepairedAt: time.Time{},
		DeletedAt:  time.Time{},

		GpuID:        "",
		SshPublicKey: params.SshPublicKey,
		Ports:        ports,
		Activities:   []string{ActivityBeingCreated},
		Subsystems:   Subsystems{},
		Specs: Specs{
			CpuCores: params.CpuCores,
			RAM:      params.RAM,
			DiskSize: params.DiskSize,
		},

		StatusCode:    status_codes.ResourceBeingCreated,
		StatusMessage: status_codes.GetMsg(status_codes.ResourceBeingCreated),
	}

	filter := bson.D{{"name", params.Name}, {"deletedAt", bson.D{{"$in", []interface{}{time.Time{}, nil}}}}}
	result, err := client.Collection.UpdateOne(context.TODO(), filter, bson.D{
		{"$setOnInsert", vm},
	}, options.Update().SetUpsert(true))
	if err != nil {
		return false, fmt.Errorf("failed to create vm. details: %w", err)
	}

	if result.UpsertedCount == 0 {
		if result.MatchedCount == 1 {
			fetchedVm, err := client.GetByName(params.Name)
			if err != nil {
				return false, err
			}

			if fetchedVm == nil {
				log.Println(fmt.Errorf("failed to fetch vm %s after creation. assuming it was deleted", params.Name))
				return false, nil
			}

			if fetchedVm.ID == vmID {
				return true, nil
			}
		}

		return false, nil
	}

	return true, nil
}

func (client *Client) GetByOwnerID(ownerID string) ([]VM, error) {
	return models.GetManyResources[VM](client.Collection, bson.D{{"ownerId", ownerID}}, false)
}

func (client *Client) CountByOwnerID(ownerID string) (int, error) {
	count, err := models.CountResources(client.Collection, bson.D{{"ownerId", ownerID}}, false)
	if err != nil {
		return 0, err
	}

	return count, nil
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

func (client *Client) UpdateWithParamsByID(id string, update *UpdateParams) error {
	updateData := bson.M{}

	models.AddIfNotNil(updateData, "ports", update.Ports)
	models.AddIfNotNil(updateData, "specs.cpuCores", update.CpuCores)
	models.AddIfNotNil(updateData, "specs.ram", update.RAM)

	if len(updateData) == 0 {
		return nil
	}

	_, err := client.Collection.UpdateOne(context.TODO(),
		bson.D{{"id", id}},
		bson.D{{"$set", updateData}},
	)
	if err != nil {
		return fmt.Errorf("failed to update vm %s. details: %w", id, err)
	}

	return nil
}

func (client *Client) UpdateSubsystemByName(name, subsystem string, key string, update interface{}) error {
	subsystemKey := fmt.Sprintf("subsystems.%s.%s", subsystem, key)
	return client.UpdateWithBsonByName(name, bson.D{{subsystemKey, update}})
}

func (client *Client) GetWithGPU() ([]VM, error) {
	// create a filter that checks if the gpuID field is not empty
	filter := bson.D{{
		"gpuId", bson.M{
			"$ne": "",
		},
	}}

	return models.GetManyResources[VM](client.Collection, filter, false)
}

func (client *Client) MarkRepaired(vmID string) error {
	filter := bson.D{{"id", vmID}}
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

func (client *Client) MarkUpdated(vmID string) error {
	filter := bson.D{{"id", vmID}}
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
