package vm_service

import (
	vmModel "go-deploy/models/vm"
	"go.mongodb.org/mongo-driver/bson"
	"log"
)

func Create(vmID, name, owner string) {
	go func() {
		err := vmModel.Create(vmID, name, owner)
		if err != nil {
			log.Println(err)
			return
		}

		err = CreateCS(name)
		if err != nil {
			log.Println(err)
		}

	}()
}

func GetByID(userID, vmID string) (*vmModel.VM, error) {
	vm, err := vmModel.GetByID(vmID)
	if err != nil {
		return nil, err
	}

	if vm != nil && vm.Owner != userID {
		return nil, nil
	}

	return vm, nil
}

func GetByName(userId, name string) (*vmModel.VM, error) {
	vm, err := vmModel.GetByName(name)
	if err != nil {
		return nil, err
	}

	if vm != nil && vm.Owner != userId {
		return nil, nil
	}

	return vm, nil
}

func GetByOwnerID(ownerID string) ([]vmModel.VM, error) {
	return vmModel.GetByOwnerID(ownerID)
}

func GetAll() ([]vmModel.VM, error) {
	return vmModel.GetAll()
}

func Exists(name string) (bool, *vmModel.VM, error) {
	return vmModel.Exists(name)
}

func MarkBeingDeleted(vmID string) error {
	return vmModel.UpdateByName(vmID, bson.D{{
		"beingDeleted", true,
	}})
}

func Delete(name string) {
	go func() {
	}()
}
