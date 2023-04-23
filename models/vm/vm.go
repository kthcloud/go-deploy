package vm

import (
	"context"
	"fmt"
	"go-deploy/models"
	"go-deploy/models/dto"
	"go-deploy/pkg/status_codes"
	csModels "go-deploy/pkg/subsystems/cs/models"
	psModels "go-deploy/pkg/subsystems/pfsense/models"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type VM struct {
	ID            string     `bson:"id"`
	Name          string     `bson:"name"`
	SshPublicKey  string     `bson:"sshPublicKey"`
	OwnerID       string     `bson:"ownerId"`
	BeingCreated  bool       `bson:"beingCreated"`
	BeingDeleted  bool       `bson:"beingDeleted"`
	Subsystems    Subsystems `bson:"subsystems"`
	StatusCode    int        `bson:"statusCode"`
	StatusMessage string     `bson:"statusMessage"`
}

type Subsystems struct {
	CS      CS      `bson:"cs"`
	PfSense PfSense `bson:"pfSense"`
}

type CS struct {
	VM                 csModels.VmPublic                 `bson:"vm"`
	PortForwardingRule csModels.PortForwardingRulePublic `bson:"portForwardingRule"`
	PublicIpAddress    csModels.PublicIpAddressPublic    `bson:"publicIpAddress"`
	SshPublicKey       string                            `bson:"sshPublicKey"`
}

type PfSense struct {
	PortForwardingRule psModels.PortForwardingRulePublic `bson:"portForwardingRule"`
}

func (vm *VM) ToDto(status, connectionString string) dto.VmRead {
	return dto.VmRead{
		ID:               vm.ID,
		Name:             vm.Name,
		SshPublicKey:     vm.SshPublicKey,
		OwnerID:          vm.OwnerID,
		Status:           status,
		ConnectionString: connectionString,
	}
}

func Create(vmID, name, sshPublicKey, owner string) error {
	currentVM, err := GetByID(vmID)
	if err != nil {
		return err
	}

	if currentVM != nil {
		return nil
	}

	vm := VM{
		ID:            vmID,
		Name:          name,
		SshPublicKey:  sshPublicKey,
		OwnerID:       owner,
		BeingCreated:  true,
		BeingDeleted:  false,
		Subsystems:    Subsystems{},
		StatusCode:    status_codes.ResourceBeingCreated,
		StatusMessage: status_codes.GetMsg(status_codes.ResourceBeingCreated),
	}

	_, err = models.VmCollection.InsertOne(context.TODO(), vm)
	if err != nil {
		err = fmt.Errorf("failed to create vm %s. details: %s", name, err)
		return err
	}

	return nil
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

func UpdateByID(id string, update bson.D) error {
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
