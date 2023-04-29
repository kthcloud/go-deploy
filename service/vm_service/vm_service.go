package vm_service

import (
	"fmt"
	vmModel "go-deploy/models/vm"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/status_codes"
	"go-deploy/service/vm_service/internal_service"
	"log"

	"go.mongodb.org/mongo-driver/bson"
)

func Create(vmID, name, sshPublicKey, owner string) {
	go func() {
		err := vmModel.Create(vmID, name, sshPublicKey, owner)
		if err != nil {
			log.Println(err)
			return
		}

		csResult, err := internal_service.CreateCS(name, sshPublicKey)
		if err != nil {
			log.Println(err)
			return
		}

		_, err = internal_service.CreatePfSense(name, csResult.PublicIpAddress.IpAddress)
		if err != nil {
			log.Println(err)
			return
		}

	}()
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

func Delete(name string) {
	go func() {
		vm, err := vmModel.GetByName(name)
		if err != nil {
			log.Println(err)
			return
		}

		err = internal_service.DeleteCS(name)
		if err != nil {
			log.Println(err)
			return
		}

		detached, err := vmModel.DetachGPU(vm.ID, vm.OwnerID)
		if err != nil {
			log.Println(err)
			return
		}

		if !detached {
			log.Println("gpu was not detached from vm", vm.ID)
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
	domainName := conf.Env.VM.ParentDomain
	port := vm.Subsystems.PfSense.PortForwardingRule.ExternalPort

	connectionString := fmt.Sprintf("ssh cloud@%s -p %d", domainName, port)

	return connectionString, nil
}

func GetStatus(vm *vmModel.VM) (int, string, error) {
	csStatusCode, csStatusMsg, err := internal_service.GetStatusCS(vm.Name)

	if err != nil {
		log.Println(err)
	}

	if err != nil || csStatusCode == status_codes.ResourceUnknown || csStatusCode == status_codes.ResourceNotFound {
		if vm.BeingDeleted {
			return status_codes.ResourceNotReady, status_codes.GetMsg(status_codes.ResourceNotReady), nil
		}

		if vm.BeingCreated {
			return status_codes.ResourceNotReady, status_codes.GetMsg(status_codes.ResourceNotReady), nil
		}
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
