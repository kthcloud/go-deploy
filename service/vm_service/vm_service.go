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

	err = vmModel.AddActivity(vm.ID, vmModel.ActivityBeingDeleted)
	if err != nil {
		return makeError(err)
	}

	err = internal_service.DeleteCS(name)
	if err != nil {
		return makeError(err)
	}

	err = gpu.Detach(vm.ID, vm.OwnerID)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func Update(vmID string, dtoVmUpdate *body.VmUpdate) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update vm. details: %s", err)
	}

	vmUpdate := &vmModel.UpdateParams{}
	vmUpdate.FromDTO(dtoVmUpdate)

	err := internal_service.UpdateCS(vmID, vmUpdate)
	if err != nil {
		return makeError(err)
	}

	err = vmModel.UpdateByID(vmID, vmUpdate)
	if err != nil {
		return makeError(err)
	}

	err = vmModel.RemoveActivity(vmID, vmModel.ActivityBeingUpdated)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func GetConnectionString(vm *vmModel.VM) (*string, error) {
	domainName := conf.Env.VM.ParentDomain
	port := vm.Subsystems.CS.PortForwardingRuleMap["__ssh"].PublicPort

	if domainName == "" || port == 0 {
		return nil, nil
	}

	connectionString := fmt.Sprintf("ssh cloud@%s -p %d", domainName, port)

	return &connectionString, nil
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
			vmModel.ActivityAttachingGPU,
			vmModel.ActivityDetachingGPU,
		}) {
			return false, "It should not be in creation, deletion, and should not be attaching or detaching a GPU", nil
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

func CheckQuotaCreate(userID string, quota *user.Quota, createParams body.VmCreate) (bool, string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check quota. details: %s", err)
	}

	usage, err := GetUsageByUserID(userID)
	if err != nil {
		return false, "", makeError(err)
	}

	totalCpuCores := usage.CpuCores + createParams.CpuCores
	totalRam := usage.RAM + createParams.RAM
	totalDiskSize := usage.DiskSize + createParams.DiskSize

	if totalCpuCores > quota.CpuCores {
		return false, fmt.Sprintf("CPU cores quota exceeded. Current: %d, Quota: %d", totalCpuCores, quota.CpuCores), nil
	}

	if totalRam > quota.RAM {
		return false, fmt.Sprintf("RAM quota exceeded. Current: %d, Quota: %d", totalRam, quota.RAM), nil
	}

	if totalDiskSize > quota.DiskSize {
		return false, fmt.Sprintf("Disk size quota exceeded. Current: %d, Quota: %d", totalDiskSize, quota.DiskSize), nil
	}

	return true, "", nil
}

func CheckQuotaUpdate(userID, vmID string, quota *user.Quota, updateParams body.VmUpdate) (bool, string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check quota. details: %s", err)
	}

	usage, err := GetUsageByUserID(userID)
	if err != nil {
		return false, "", makeError(err)
	}

	if usage == nil {
		return false, "", makeError(fmt.Errorf("usage not found"))
	}

	vm, err := vmModel.GetByID(vmID)
	if err != nil {
		return false, "", makeError(err)
	}

	if vm == nil {
		return false, "", makeError(fmt.Errorf("vm not found"))
	}

	totalCpuCores := usage.CpuCores
	if updateParams.CpuCores != nil {
		totalCpuCores += *updateParams.CpuCores - vm.Specs.CpuCores
	}

	totalRam := usage.RAM
	if updateParams.RAM != nil {
		totalRam += *updateParams.RAM - vm.Specs.RAM
	}

	if totalCpuCores > quota.CpuCores {
		return false, fmt.Sprintf("CPU cores quota exceeded. Current: %d, Quota: %d", totalCpuCores, quota.CpuCores), nil
	}

	if totalRam > quota.RAM {
		return false, fmt.Sprintf("RAM quota exceeded. Current: %d, Quota: %d", totalRam, quota.RAM), nil
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
		specs := vm.Specs

		usage.CpuCores += specs.CpuCores
		usage.RAM += specs.RAM
		usage.DiskSize += specs.DiskSize
	}

	return usage, nil
}

func GetExternalPortMapper(vmID string) (map[string]int, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get external port mapper. details: %s", err)
	}

	vm, err := vmModel.GetByID(vmID)
	if err != nil {
		return nil, makeError(err)
	}

	if vm == nil {
		return nil, nil
	}

	mapper := make(map[string]int)
	for name, port := range vm.Subsystems.CS.PortForwardingRuleMap {
		mapper[name] = port.PublicPort
	}

	return mapper, nil
}
