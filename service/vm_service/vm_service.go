package vm_service

import (
	"fmt"
	"go-deploy/models/dto/body"
	"go-deploy/models/sys/user"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/models/sys/vm/gpu"
	"go-deploy/pkg/conf"
	"go-deploy/service/vm_service/internal_service"
	"log"
)

func Create(vmID, owner string, vmCreate *body.VmCreate) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create vm. details: %s", err)
	}

	params := &vmModel.CreateParams{}
	params.FromDTO(vmCreate)

	err := vmModel.Create(vmID, owner, conf.Env.Manager, params)
	if err != nil {
		return makeError(err)
	}

	_, err = internal_service.CreateCS(params)
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

func Delete(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete vm. details: %s", err)
	}

	vm, err := vmModel.GetByName(name)
	if err != nil {
		return makeError(err)
	}

	if vm == nil {
		return nil
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

	return nil
}

func Update(vmID string, dtoVmUpdate *body.VmUpdate) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update vm. details: %s", err)
	}

	vmUpdate := vmModel.UpdateParams{}
	vmUpdate.FromDTO(dtoVmUpdate)

	err := internal_service.UpdateCS(vmID, vmUpdate.Ports)
	if err != nil {
		return makeError(err)
	}

	err = vmModel.UpdateByID(vmID, &vmUpdate)
	if err != nil {
		return makeError(err)
	}

	err = vmModel.RemoveActivity(vmID, vmModel.ActivityBeingUpdated)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func GetConnectionString(vm *vmModel.VM) (string, error) {
	domainName := conf.Env.VM.ParentDomain
	port := vm.Subsystems.CS.PortForwardingRuleMap["__ssh"].PublicPort

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

func StartActivity(vmID, activity string) (bool, string, error) {
	canAdd, reason, err := CanAddActivity(vmID, activity)
	if err != nil {
		return false, "", err
	}

	if !canAdd {
		return false, reason, nil
	}

	err = vmModel.AddActivity(vmID, activity)
	if err != nil {
		return false, "", err
	}

	return true, "", nil
}

func CanAddActivity(deploymentID, activity string) (bool, string, error) {
	vm, err := vmModel.GetByID(deploymentID)
	if err != nil {
		return false, "", err
	}

	if vm == nil {
		return false, "", err
	}

	switch activity {
	case vmModel.ActivityBeingCreated:
		return !vm.BeingDeleted(), "It is being deleted", nil
	case vmModel.ActivityBeingDeleted:
		return !vm.BeingCreated(), "It is being created", nil
	case vmModel.ActivityBeingUpdated:
		if vm.DoingOnOfActivities([]string{
			vmModel.ActivityBeingCreated,
			vmModel.ActivityBeingDeleted,
			vmModel.ActivityBeingUpdated,
			vmModel.ActivityAttachingGPU,
			vmModel.ActivityDetachingGPU,
		}) {
			return false, "It should not be in creation, deletion, or updating, and should not be attaching or detaching a GPU", nil
		}
		return true, "", nil
	case vmModel.ActivityAttachingGPU:
		if vm.DoingOnOfActivities([]string{
			vmModel.ActivityBeingCreated,
			vmModel.ActivityBeingDeleted,
			vmModel.ActivityAttachingGPU,
			vmModel.ActivityDetachingGPU,
		}) {
			return false, "It should not be in creation or deletion, and should not be attaching or detaching a GPU", nil
		}
		return true, "", nil
	case vmModel.ActivityDetachingGPU:
		if vm.DoingOnOfActivities([]string{
			vmModel.ActivityBeingCreated,
			vmModel.ActivityBeingDeleted,
			vmModel.ActivityAttachingGPU,
		}) {
			return false, "It should not be in creation or deletion, and should not be attaching a GPU", nil
		}
		return true, "", nil
	}

	return false, "", fmt.Errorf("unknown activity %s", activity)
}

func CheckQuota(userID string, quota *user.Quota, createParams body.VmCreate) (bool, string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check quota. details: %s", err)
	}

	usage, err := GetUsageByUserID(userID)
	if err != nil {
		return false, "", makeError(err)
	}

	totalCpuCores := usage.CpuCores + createParams.CpuCores
	totalRam := usage.RAM + createParams.RAM
	totalDiskSize := usage.DiskSpace + createParams.DiskSize

	if totalCpuCores > quota.CpuCores {
		return false, fmt.Sprintf("CPU cores quota exceeded. Current: %d, Quota: %d", totalCpuCores, quota.CpuCores), nil
	}

	if totalRam > quota.RAM {
		return false, fmt.Sprintf("RAM quota exceeded. Current: %d, Quota: %d", totalRam, quota.RAM), nil
	}

	if totalDiskSize > quota.DiskSpace {
		return false, fmt.Sprintf("Disk space quota exceeded. Current: %d, Quota: %d", totalDiskSize, quota.DiskSpace), nil
	}

	return true, "", nil
}

func GetUsageByUserID(id string) (*vmModel.Usage, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check quota. details: %s", err)
	}

	usage := &vmModel.Usage{}

	currentVms, err := vmModel.GetByOwnerID(id)
	if err != nil {
		return nil, makeError(err)
	}

	for _, vm := range currentVms {
		so := vm.Subsystems.CS.ServiceOffering

		usage.CpuCores += so.CpuCores
		usage.RAM += so.RAM
		usage.DiskSpace += so.DiskSize
	}

	return usage, nil
}
