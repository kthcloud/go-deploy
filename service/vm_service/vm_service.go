package vm_service

import (
	"fmt"
	"go-deploy/models/dto/body"
	"go-deploy/models/sys/vm"
	"go-deploy/models/sys/vm/gpu"
	"go-deploy/pkg/conf"
	"go-deploy/service/vm_service/internal_service"
	"log"

	"go.mongodb.org/mongo-driver/bson"
)

func Create(vmID, name, sshPublicKey, owner string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create vm. details: %s", err)
	}

	err := vm.Create(vmID, name, sshPublicKey, owner, conf.Env.Manager)
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

func GetByID(userID, vmID string, isAdmin bool) (*vm.VM, error) {
	vm, err := vm.GetByID(vmID)
	if err != nil {
		return nil, err
	}

	if vm != nil && vm.OwnerID != userID && !isAdmin {
		return nil, nil
	}

	return vm, nil
}

func GetByOwnerID(ownerID string) ([]vm.VM, error) {
	return vm.GetByOwnerID(ownerID)
}

func GetAll() ([]vm.VM, error) {
	return vm.GetAll()
}

func GetCount(userID string) (int, error) {
	return vm.CountByOwnerID(userID)
}

func Exists(name string) (bool, *vm.VM, error) {
	return vm.Exists(name)
}

func MarkBeingDeleted(vmID string) error {
	return vm.UpdateWithBsonByID(vmID, bson.D{{
		"beingDeleted", true,
	}})
}

func Delete(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete vm. details: %s", err)
	}

	vm, err := vm.GetByName(name)
	if err != nil {
		return makeError(err)
	}

	err = internal_service.DeleteCS(name)
	if err != nil {
		return makeError(err)
	}

	detached, err := gpu.DetachGPU(vm.ID, vm.OwnerID)
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

func Update(vmID string, dtoVmUpdate *body.VmUpdate) error {
	vmUpdate := vm.VmUpdate{}

	if dtoVmUpdate.Ports != nil {
		var ports []vm.Port
		for _, port := range *dtoVmUpdate.Ports {
			ports = append(ports, vm.Port{
				Name:     port.Name,
				Port:     port.Port,
				Protocol: port.Protocol,
			})
		}
		vmUpdate.Ports = &ports
	}

	err := vm.UpdateByID(vmID, &vmUpdate)
	if err != nil {
		return fmt.Errorf("failed to update vm. details: %s", err)
	}

	return nil
}

func GetConnectionString(vm *vm.VM) (string, error) {
	domainName := conf.Env.VM.ParentDomain
	port := vm.Subsystems.PfSense.PortForwardingRuleMap["ssh"].ExternalPort

	connectionString := fmt.Sprintf("ssh cloud@%s -p %d", domainName, port)

	return connectionString, nil
}

func DoCommand(vm *vm.VM, command string) {
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
