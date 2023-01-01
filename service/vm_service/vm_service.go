package vm_service

import (
	"errors"
	"fmt"
	vmModel "go-deploy/models/vm"
	"go-deploy/pkg/conf"
	"go-deploy/service/vm_service/internal_service"
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

		csResult, err := internal_service.CreateCS(name)
		if err != nil {
			log.Println(err)
		}

		_, err = internal_service.CreatePfSense(name, csResult.PublicIpAddress.IpAddress)
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
	return vmModel.UpdateByID(vmID, bson.D{{
		"beingDeleted", true,
	}})
}

func Delete(name string) {
	go func() {
		err := internal_service.DeleteCS(name)
		if err != nil {
			log.Println(err)
			return
		}

		err = internal_service.DeletePfSense(name)
		if err != nil {
			log.Println(err)
			return
		}
	}()
}

func GetConnectionStringByID(vmID string) (string, error) {
	vm, err := vmModel.GetByID(vmID)
	if err != nil {
		return "", err
	}

	domainName := conf.Env.ParentDomainVM
	port := vm.Subsystems.PfSense.PortForwardingRule.ExternalPort

	connectionString := fmt.Sprintf("ssh cloud@%s -p %d", domainName, port)

	return connectionString, nil
}

func CreateKeyPairByID(vmID, publicKey string) error {
	vm, err := vmModel.GetByID(vmID)
	if err != nil {
		return err
	}

	csID := vm.Subsystems.CS.VM.ID
	if csID == "" {
		return errors.New("cloudstack vm not created")
	}

	err = internal_service.AddKeyPairCS(csID, publicKey)
	return err
}
