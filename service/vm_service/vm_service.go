package vm_service

import (
	"fmt"
	roleModel "go-deploy/models/config/role"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/query"
	teamModels "go-deploy/models/sys/user/team"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/models/sys/vm/gpu"
	"go-deploy/pkg/config"
	"go-deploy/service"
	"go-deploy/service/vm_service/cs_service"
	"go-deploy/service/vm_service/k8s_service"
	"go-deploy/utils"
	"log"
	"sort"
	"strings"
)

func Create(id, ownerID string, vmCreate *body.VmCreate) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create vm. details: %w", err)
	}

	// temporary hard-coded fallback
	fallback := "se-flem"
	deploymentZone := "se-flem"

	params := &vmModel.CreateParams{}
	params.FromDTO(vmCreate, &fallback, &deploymentZone)

	created, err := vmModel.New().Create(id, ownerID, config.Config.Manager, params)
	if err != nil {
		return makeError(err)
	}

	if !created {
		return makeError(fmt.Errorf("vm already exists for another user"))
	}

	err = cs_service.CreateCS(id, params)
	if err != nil {
		return makeError(err)
	}

	// there is always at least one port: __ssh
	if len(params.Ports) > 1 {
		err = k8s_service.Create(id, params)
		if err != nil {
			return makeError(err)
		}
	} else {
		log.Println("skipping k8s setup for vm", id, "since it has no ports")
	}

	return nil
}

func Update(id string, dtoVmUpdate *body.VmUpdate) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update vm. details: %w", err)
	}

	vmUpdate := &vmModel.UpdateParams{}
	vmUpdate.FromDTO(dtoVmUpdate)

	if vmUpdate.SnapshotID != nil {
		err := ApplySnapshot(id, *vmUpdate.SnapshotID)
		if err != nil {
			return makeError(err)
		}
	} else {
		err := vmModel.New().UpdateWithParamsByID(id, vmUpdate)
		if err != nil {
			return makeError(err)
		}

		err = cs_service.UpdateCS(id, vmUpdate)
		if err != nil {
			return makeError(err)
		}

		if vmUpdate.Ports != nil {
			err = k8s_service.Repair(id)
			if err != nil {
				return makeError(err)
			}
		}
	}

	err := vmModel.New().RemoveActivity(id, vmModel.ActivityBeingUpdated)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func Delete(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete vm. details: %w", err)
	}

	vm, err := vmModel.New().GetByID(id)
	if err != nil {
		return makeError(err)
	}

	if vm == nil {
		return nil
	}

	err = vmModel.New().AddActivity(vm.ID, vmModel.ActivityBeingDeleted)
	if err != nil {
		return makeError(err)
	}

	err = k8s_service.Delete(id)
	if err != nil {
		return makeError(err)
	}

	err = cs_service.DeleteCS(vm.ID)
	if err != nil {
		return makeError(err)
	}

	err = gpu.New().Detach(vm.ID, vm.OwnerID)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func Repair(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to repair vm %s. details: %w", id, err)
	}

	vm, err := vmModel.New().GetByID(id)
	if err != nil {
		return makeError(err)
	}

	if vm == nil {
		log.Println("vm", id, "not found when repairing. assuming it was deleted")
		return nil
	}

	if !vm.Ready() {
		log.Println("vm", id, "not ready when repairing.")
		return nil
	}

	started, reason, err := StartActivity(vm.ID, vmModel.ActivityRepairing)
	if err != nil {
		return makeError(err)
	}

	if !started {
		return fmt.Errorf("failed to repair vm. details: %s", reason)
	}

	defer func() {
		err = vmModel.New().RemoveActivity(vm.ID, vmModel.ActivityRepairing)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to remove activity %s from vm %s details: %w", vmModel.ActivityRepairing, vm.Name, err))
		}
	}()

	err = cs_service.RepairCS(id)
	if err != nil {
		return makeError(err)
	}

	log.Println("successfully repaired vm", id)
	return nil
}

func GetByIdAuth(id string, auth *service.AuthInfo) (*vmModel.VM, error) {
	vm, err := vmModel.New().GetByID(id)
	if err != nil {
		return nil, err
	}

	if vm == nil {
		return nil, nil
	}

	if vm.OwnerID != auth.UserID {
		inTeam, err := teamModels.New().AddUserID(auth.UserID).AddResourceID(id).ExistsAny()
		if err != nil {
			return nil, err
		}

		if inTeam {
			return vm, nil
		}

		if auth.IsAdmin {
			return vm, nil
		}

		return nil, nil
	}

	return vm, nil
}

func GetByID(id string) (*vmModel.VM, error) {
	return vmModel.New().GetByID(id)
}

func GetByName(name string) (*vmModel.VM, error) {
	return vmModel.New().GetByName(name)
}

func ListAuth(allUsers bool, userID *string, shared bool, auth *service.AuthInfo, pagination *query.Pagination) ([]vmModel.VM, error) {
	client := vmModel.New()

	if pagination != nil {
		client.AddPagination(pagination.Page, pagination.PageSize)
	}

	if userID != nil {
		if *userID != auth.UserID && !auth.IsAdmin {
			return nil, nil
		}
		client.RestrictToOwner(*userID)
	} else if !allUsers || (allUsers && !auth.IsAdmin) {
		client.RestrictToOwner(auth.UserID)
	}

	resources, err := client.ListAll()
	if err != nil {
		return nil, err
	}

	if shared {
		ids := make([]string, len(resources))
		for i, resource := range resources {
			ids[i] = resource.ID
		}

		teamClient := teamModels.New().AddUserID(auth.UserID)
		if pagination != nil {
			teamClient.AddPagination(pagination.Page, pagination.PageSize)
		}

		teams, err := teamClient.ListAll()
		if err != nil {
			return nil, err
		}

		for _, team := range teams {
			for _, resource := range team.GetResourceMap() {
				if resource.Type != teamModels.ResourceTypeVM {
					continue
				}

				// skip existing non-shared resources
				skip := false
				for _, id := range ids {
					if resource.ID == id {
						skip = true
						break
					}
				}
				if skip {
					continue
				}

				vm, err := vmModel.New().GetByID(resource.ID)
				if err != nil {
					return nil, err
				}

				if vm != nil {
					resources = append(resources, *vm)
				}
			}
		}

		sort.Slice(resources, func(i, j int) bool {
			return resources[i].CreatedAt.After(resources[j].CreatedAt)
		})

		// since we fetched from two collections, we need to do pagination manually
		if pagination != nil {
			resources = utils.GetPage(resources, pagination.PageSize, pagination.Page)
		}

	} else {
		// sort by createdAt
		sort.Slice(resources, func(i, j int) bool {
			return resources[i].CreatedAt.After(resources[j].CreatedAt)
		})
	}

	return resources, nil
}

func GetConnectionString(vm *vmModel.VM) (*string, error) {
	if vm == nil {
		return nil, nil
	}

	zone := config.Config.VM.GetZone(vm.Zone)
	if zone == nil {
		return nil, fmt.Errorf("zone %s not found", vm.Zone)
	}

	domainName := zone.ParentDomain
	port := vm.Subsystems.CS.PortForwardingRuleMap["__ssh"].PublicPort

	if domainName == "" || port == 0 {
		return nil, nil
	}

	connectionString := fmt.Sprintf("ssh root@%s -p %d", domainName, port)

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
		if vm.HasGPU() {
			gpuID = &vm.GpuID
		}

		err := cs_service.DoCommandCS(csID, gpuID, command, vm.Zone)
		if err != nil {
			utils.PrettyPrintError(err)
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

	err = vmModel.New().AddActivity(vmID, activity)
	if err != nil {
		return false, "", err
	}

	return true, "", nil
}

func CanAddActivity(vmID, activity string) (bool, string, error) {
	vm, err := vmModel.New().GetByID(vmID)
	if err != nil {
		return false, "", err
	}

	if vm == nil {
		return false, "", err
	}

	switch activity {
	case vmModel.ActivityBeingCreated:
		return !vm.BeingDeleted(), "Resource is being deleted", nil
	case vmModel.ActivityBeingDeleted:
		return true, "", nil
	case vmModel.ActivityBeingUpdated:
		if vm.DoingOnOfActivities([]string{
			vmModel.ActivityBeingCreated,
			vmModel.ActivityBeingDeleted,
			vmModel.ActivityAttachingGPU,
			vmModel.ActivityDetachingGPU,
		}) {
			return false, "Resource should not be in creation, deletion, and should not be attaching or detaching a GPU", nil
		}
		return true, "", nil
	case vmModel.ActivityAttachingGPU:
		if vm.DoingOnOfActivities([]string{
			vmModel.ActivityBeingCreated,
			vmModel.ActivityBeingDeleted,
			vmModel.ActivityAttachingGPU,
			vmModel.ActivityDetachingGPU,
			vmModel.ActivityCreatingSnapshot,
			vmModel.ActivityApplyingSnapshot,
		}) {
			return false, "Resource should not be in creation or deletion, and should not be attaching or detaching a GPU", nil
		}
		return true, "", nil
	case vmModel.ActivityDetachingGPU:
		if vm.DoingOnOfActivities([]string{
			vmModel.ActivityBeingCreated,
			vmModel.ActivityBeingDeleted,
			vmModel.ActivityAttachingGPU,
			vmModel.ActivityCreatingSnapshot,
			vmModel.ActivityApplyingSnapshot,
		}) {
			return false, "Resource should not be in creation or deletion, and should not be attaching a GPU", nil
		}
		return true, "", nil
	case vmModel.ActivityRepairing:
		if vm.DoingOnOfActivities([]string{
			vmModel.ActivityBeingCreated,
			vmModel.ActivityBeingDeleted,
			vmModel.ActivityAttachingGPU,
			vmModel.ActivityDetachingGPU,
		}) {
			return false, "Resource should not be in creation or deletion, and should not be attaching or detaching a GPU", nil
		}
		return true, "", nil
	case vmModel.ActivityCreatingSnapshot:
		if vm.DoingOnOfActivities([]string{
			vmModel.ActivityBeingCreated,
			vmModel.ActivityBeingDeleted,
			vmModel.ActivityAttachingGPU,
			vmModel.ActivityDetachingGPU,
		}) {
			return false, "Resource should not be in creation or deletion, and should not be attaching or detaching a GPU", nil
		}
		return true, "", nil
	case vmModel.ActivityApplyingSnapshot:
		if vm.DoingOnOfActivities([]string{
			vmModel.ActivityBeingCreated,
			vmModel.ActivityBeingDeleted,
			vmModel.ActivityAttachingGPU,
			vmModel.ActivityDetachingGPU,
		}) {
			return false, "Resource should not be in creation or deletion, and should not be attaching or detaching a GPU", nil
		}
		return true, "", nil
	}

	return false, "", fmt.Errorf("unknown activity %s", activity)
}

func CheckQuotaCreate(userID string, quota *roleModel.Quotas, auth *service.AuthInfo, createParams body.VmCreate) (bool, string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check quota. details: %w", err)
	}

	if auth.IsAdmin {
		return true, "", nil
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

func CheckQuotaUpdate(userID, vmID string, quota *roleModel.Quotas, auth *service.AuthInfo, updateParams body.VmUpdate) (bool, string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check quota. details: %w", err)
	}

	if auth.IsAdmin {
		return true, "", nil
	}

	if updateParams.CpuCores == nil && updateParams.RAM == nil {
		return true, "", nil
	}

	usage, err := GetUsageByUserID(userID)
	if err != nil {
		return false, "", makeError(err)
	}

	if usage == nil {
		return false, "", makeError(fmt.Errorf("usage not found"))
	}

	vm, err := vmModel.New().GetByID(vmID)
	if err != nil {
		return false, "", makeError(err)
	}

	if vm == nil {
		return false, "", makeError(fmt.Errorf("vm not found"))
	}

	if updateParams.CpuCores != nil {
		totalCpuCores := usage.CpuCores
		if updateParams.CpuCores != nil {
			totalCpuCores += *updateParams.CpuCores - vm.Specs.CpuCores
		}

		if totalCpuCores > quota.CpuCores {
			return false, fmt.Sprintf("CPU cores quota exceeded. Current: %d, Quota: %d", totalCpuCores, quota.CpuCores), nil
		}
	}

	if updateParams.RAM != nil {
		totalRam := usage.RAM
		if updateParams.RAM != nil {
			totalRam += *updateParams.RAM - vm.Specs.RAM
		}

		if totalRam > quota.RAM {
			return false, fmt.Sprintf("RAM quota exceeded. Current: %d, Quota: %d", totalRam, quota.RAM), nil
		}
	}

	return true, "", nil
}

func GetUsageByUserID(id string) (*vmModel.Usage, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check quota. details: %w", err)
	}

	usage := &vmModel.Usage{}

	currentVms, err := vmModel.New().RestrictToOwner(id).ListAll()
	if err != nil {
		return nil, makeError(err)
	}

	for _, vm := range currentVms {
		specs := vm.Specs

		usage.CpuCores += specs.CpuCores
		usage.RAM += specs.RAM
		usage.DiskSize += specs.DiskSize

		for _, snapshot := range vm.Subsystems.CS.SnapshotMap {
			if strings.Contains(snapshot.Description, "user") {
				usage.Snapshots++
			}
		}
	}

	return usage, nil
}

func GetExternalPortMapper(vmID string) (map[string]int, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get external port mapper. details: %w", err)
	}

	vm, err := vmModel.New().GetByID(vmID)
	if err != nil {
		return nil, makeError(err)
	}

	if vm == nil {
		log.Println("vm", vmID, "not found for when detaching getting external port mapper. assuming it was deleted")
		return nil, nil
	}

	mapper := make(map[string]int)
	for name, port := range vm.Subsystems.CS.PortForwardingRuleMap {
		mapper[name] = port.PublicPort
	}

	return mapper, nil
}

func removeDeploySshFromPortMap(portMap *[]vmModel.Port) {
	for i, port := range *portMap {
		if (port.Port == 22 || port.Name == "__ssh") && port.Protocol == "tcp" {
			*portMap = append((*portMap)[:i], (*portMap)[i+1:]...)
			break
		}
	}
}
