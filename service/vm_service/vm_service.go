package vm_service

import (
	"fmt"
	vmModel "go-deploy/models/vm"
	"go-deploy/pkg/conf"
	"go-deploy/service/vm_service/internal_service"
	"log"

	"go.mongodb.org/mongo-driver/bson"
)

func Create(vmID, name, sshPublicKey, owner string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create vm. details: %s", err)
	}

	err := vmModel.Create(vmID, name, sshPublicKey, owner, conf.Env.Manager)
	if err != nil {
		return makeError(err)
	}

	csResult, err := internal_service.CreateCS(name, sshPublicKey)
	if err != nil {
		return makeError(err)
	}

	_, err = internal_service.CreatePfSense(name, csResult.PublicIpAddress.IpAddress)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func GetByID(userID, vmID string, isAdmin bool) (*vmModel.VM, error) {
	vm, err := vmModel.GetByID(vmID)
	if err != nil {
		return nil, err
	}

	if vm != nil && vm.OwnerID != userID && !isAdmin {
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

func GetCount(userID string) (int, error) {
	return vmModel.CountByOwnerID(userID)
}

func Exists(name string) (bool, *vmModel.VM, error) {
	return vmModel.Exists(name)
}

func MarkBeingDeleted(vmID string) error {
	return vmModel.UpdateByID(vmID, bson.D{{
		"beingDeleted", true,
	}})
}

func Delete(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete vm. details: %s", err)
	}

	vm, err := vmModel.GetByName(name)
	if err != nil {
		return makeError(err)
	}

	err = internal_service.DeleteCS(name)
	if err != nil {
		return makeError(err)
	}

	detached, err := vmModel.DetachGPU(vm.ID, vm.OwnerID)
	if err != nil {
		return makeError(err)
	}

	if !detached {
		return makeError(fmt.Errorf("failed to detach gpu from vm"))
	}

	err = internal_service.DeletePfSense(name)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func GetConnectionString(vm *vmModel.VM) (string, error) {
	domainName := conf.Env.VM.ParentDomain
	port := vm.Subsystems.PfSense.PortForwardingRule.ExternalPort

	connectionString := fmt.Sprintf("ssh cloud@%s -p %d", domainName, port)

	return connectionString, nil
}

func DoCommand(vm *vmModel.VM, command string) {
	go func() {
		csID := vm.Subsystems.CS.VM.ID
		if csID == "" {
			log.Println("cannot execute any command when cloudstack vm is not set up")
			return
		}

		var gpuID *string
		if vm.GpuID != "" {
			gpuID = &vm.GpuID
		}

		err := internal_service.DoCommandCS(csID, gpuID, command)
		if err != nil {
			log.Println(err)
			return
		}
	}()
}
