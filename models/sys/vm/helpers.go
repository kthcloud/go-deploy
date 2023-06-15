package vm

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


func Create(vmID, owner, manager string, params *CreateParams) (bool, error) {
	vm := VM{
		ID:        vmID,
		Name:      params.Name,
		OwnerID:   owner,
		ManagedBy: manager,

		CreatedAt: time.Now(),

		GpuID:        "",
		SshPublicKey: params.SshPublicKey,
		Ports:        []Port{},

		Activities: []string{ActivityBeingCreated},
		Subsystems: Subsystems{},

		Specs: Specs{
			CpuCores: params.CpuCores,
			RAM:      params.RAM,
			DiskSize: params.DiskSize,
		},

		StatusCode:    status_codes.ResourceBeingCreated,
		StatusMessage: status_codes.GetMsg(status_codes.ResourceBeingCreated),
	}

	result, err := models.VmCollection.UpdateOne(context.TODO(), bson.D{{"name", params.Name}}, bson.D{
		{"$setOnInsert", vm},
	}, options.Update().SetUpsert(true))
	if err != nil {
		return false, fmt.Errorf("failed to create vm. details: %s", err)
	}

	if result.UpsertedCount == 0 {
		return false, nil
	}

	return true, nil
}

func getVM(filter bson.D) (*VM, error) {
	var vm VM
	err := models.VmCollection.FindOne(context.TODO(), filter).Decode(&vm)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}

		err = fmt.Errorf("failed to fetch vm. details: %s", err)
		return nil, err
	}

	return &vm, err
}

func GetByID(vmID string) (*VM, error) {
	return getVM(bson.D{{"id", vmID}})
}

func GetByName(name string) (*VM, error) {
	return getVM(bson.D{{"name", name}})
}

func Exists(name string) (bool, *VM, error) {
	vm, err := getVM(bson.D{{"name", name}})
	if err != nil {
		return false, nil, err
	}

	if vm == nil {
		return false, nil, nil
	}

	return true, vm, err
}

func GetByOwnerID(ownerID string) ([]VM, error) {
	cursor, err := models.VmCollection.Find(context.TODO(), bson.D{{"ownerId", ownerID}})

	if err != nil {
		err = fmt.Errorf("failed to find vms from owner ID %s. details: %s", ownerID, err)
		log.Println(err)
		return nil, err
	}

	var vms []VM
	for cursor.Next(context.TODO()) {
		var vm VM

		err = cursor.Decode(&vm)
		if err != nil {
			err = fmt.Errorf("failed to fetch vm when fetching all vms from owner ID %s. details: %s", ownerID, err)
			log.Println(err)
			return nil, err
		}
		vms = append(vms, vm)
	}

	return vms, nil
}

func CountByOwnerID(ownerID string) (int, error) {
	count, err := models.VmCollection.CountDocuments(context.TODO(), bson.D{{"ownerId", ownerID}})

	if err != nil {
		err = fmt.Errorf("failed to count vms by owner ID %s. details: %s", ownerID, err)
		log.Println(err)
		return 0, err
	}

	return int(count), nil
}

func DeleteByID(vmID, userID string) error {
	_, err := models.VmCollection.DeleteOne(context.TODO(), bson.D{{"id", vmID}, {"ownerId", userID}})
	if err != nil {
		err = fmt.Errorf("failed to delete vm %s. details: %s", vmID, err)
		log.Println(err)
		return err
	}
	return nil
}

func UpdateByID(id string, update *UpdateParams) error {
	updateData := bson.M{}

	models.AddIfNotNil(updateData, "ports", update.Ports)
	models.AddIfNotNil(updateData, "specs.cpuCores", update.CpuCores)
	models.AddIfNotNil(updateData, "specs.ram", update.RAM)

	if len(updateData) == 0 {
		return nil
	}

	_, err := models.VmCollection.UpdateOne(context.TODO(),
		bson.D{{"id", id}},
		bson.D{{"$set", updateData}},
	)
	if err != nil {
		return fmt.Errorf("failed to update vm %s. details: %s", id, err)
	}

	return nil
}

func UpdateWithBsonByID(id string, update bson.D) error {
	_, err := models.VmCollection.UpdateOne(context.TODO(), bson.D{{"id", id}}, bson.D{{"$set", update}})
	if err != nil {
		err = fmt.Errorf("failed to update vm %s. details: %s", id, err)
		log.Println(err)
		return err
	}
	return nil
}

func UpdateByName(name string, update bson.D) error {
	_, err := models.VmCollection.UpdateOne(context.TODO(), bson.D{{"name", name}}, bson.D{{"$set", update}})
	if err != nil {
		err = fmt.Errorf("failed to update vm %s. details: %s", name, err)
		log.Println(err)
		return err
	}
	return nil
}

func UpdateSubsystemByName(name, subsystem string, key string, update interface{}) error {
	subsystemKey := fmt.Sprintf("subsystems.%s.%s", subsystem, key)
	return UpdateByName(name, bson.D{{subsystemKey, update}})
}

func GetAll() ([]VM, error) {
	return GetAllWithFilter(bson.D{})
}

func GetAllWithFilter(filter bson.D) ([]VM, error) {
	cursor, err := models.VmCollection.Find(context.TODO(), filter)

	if err != nil {
		err = fmt.Errorf("failed to fetch all vms. details: %s", err)
		log.Println(err)
		return nil, err
	}

	var vms []VM
	for cursor.Next(context.TODO()) {
		var vm VM

		err = cursor.Decode(&vm)
		if err != nil {
			err = fmt.Errorf("failed to decode vm when fetching all vm. details: %s", err)
			log.Println(err)
			return nil, err
		}
		vms = append(vms, vm)
	}

	return vms, nil
}

func GetByActivity(activity string) ([]VM, error) {
	filter := bson.D{
		{
			"activities", bson.M{
				"$in": bson.A{activity},
			},
		},
	}

	return GetAllWithFilter(filter)
}

func GetWithNoActivities() ([]VM, error) {
	filter := bson.D{
		{
			"activities", bson.M{
				"$size": 0,
			},
		},
	}

	return GetAllWithFilter(filter)
}

func GetWithGPU() ([]VM, error) {
	// create a filter that checks if the gpuID field is not empty
	filter := bson.D{
		{
			"gpuId", bson.M{
				"$ne": "",
			},
		},
	}

	return GetAllWithFilter(filter)
}

func AddActivity(vmID, activity string) error {
	_, err := models.VmCollection.UpdateOne(context.TODO(),
		bson.D{{"id", vmID}},
		bson.D{{"$addToSet", bson.D{{"activities", activity}}}},
	)
	if err != nil {
		err = fmt.Errorf("failed to add activity %s to vm %s. details: %s", activity, vmID, err)
		return err
	}
	return nil
}

func RemoveActivity(vmID, activity string) error {
	_, err := models.VmCollection.UpdateOne(context.TODO(),
		bson.D{{"id", vmID}},
		bson.D{{"$pull", bson.D{{"activities", activity}}}},
	)
	if err != nil {
		err = fmt.Errorf("failed to remove activity %s from vm %s. details: %s", activity, vmID, err)
		return err
	}
	return nil
}

func MarkRepaired(vmID string) error {
	filter := bson.D{{"id", vmID}}
	update := bson.D{
		{"$set", bson.D{{"repairedAt", time.Now()}}},
		{"$pull", bson.D{{"activities", "repairing"}}},
	}

	_, err := models.VmCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	return nil
}

func MarkUpdated(vmID string) error {
	filter := bson.D{{"id", vmID}}
	update := bson.D{
		{"$set", bson.D{{"updatedAt", time.Now()}}},
		{"$pull", bson.D{{"activities", "updating"}}},
	}

	_, err := models.VmCollection.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		return err
	}

	return nil
}
