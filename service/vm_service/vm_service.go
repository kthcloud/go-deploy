package vm_service

import (
	"errors"
	"fmt"
	vmModel "go-deploy/models/vm"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/status_codes"
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

func GetConnectionString(vm *vmModel.VM) (string, error) {
	domainName := conf.Env.ParentDomainVM
	port := vm.Subsystems.PfSense.PortForwardingRule.ExternalPort

	connectionString := fmt.Sprintf("ssh cloud@%s -p %d", domainName, port)

	return connectionString, nil
}

func CreateKeyPair(vm *vmModel.VM, publicKey string) error {
	csID := vm.Subsystems.CS.VM.ID
	if csID == "" {
		return errors.New("cloudstack vm not created")
	}

	err := internal_service.AddKeyPairCS(csID, publicKey)
	return err
}

func GetStatus(vm *vmModel.VM) (int, string, error) {
	csStatusCode, csStatusMsg, err := internal_service.GetStatusCS(vm.Name)
	if err != nil || csStatusCode == status_codes.ResourceUnknown {
		if vm.BeingDeleted {
			return status_codes.ResourceBeingDeleted, status_codes.GetMsg(status_codes.ResourceBeingDeleted), nil
		}

		if vm.BeingCreated {
			return status_codes.ResourceBeingCreated, status_codes.GetMsg(status_codes.ResourceBeingCreated), nil
		}

		return status_codes.ResourceUnknown, status_codes.GetMsg(status_codes.ResourceUnknown), nil
	}

	return csStatusCode, csStatusMsg, nil
}

func DoCommand(vm *vmModel.VM, command string) {
	go func() {
		csID := vm.Subsystems.CS.VM.ID
		if csID == "" {
			log.Println("cannot execute any command when cloudstack vm is not set up")
			return
		}

		err := internal_service.DoCommandCS(csID, command)
		if err != nil {
			log.Println(err)
			return
		}
	}()
}
