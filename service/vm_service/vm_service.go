package vm_service

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/query"
	jobModel "go-deploy/models/sys/job"
	notificationModel "go-deploy/models/sys/notification"
	roleModel "go-deploy/models/sys/role"
	teamModels "go-deploy/models/sys/team"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/models/sys/vm/gpu"
	"go-deploy/pkg/config"
	"go-deploy/service"
	"go-deploy/service/job_service"
	"go-deploy/service/notification_service"
	"go-deploy/service/vm_service/cs_service"
	"go-deploy/service/vm_service/k8s_service"
	"go-deploy/utils"
	"log"
	"sort"
	"strings"
)

func Create(id, ownerID string, dtoVmCreate *body.VmCreate) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create vm. details: %w", err)
	}

	// temporary hard-coded fallback
	fallback := "se-flem"
	deploymentZone := "se-flem"

	params := &vmModel.CreateParams{}
	params.FromDTO(dtoVmCreate, &fallback, &deploymentZone)

	_, err := vmModel.New().Create(id, ownerID, config.Config.Manager, params)
	if err != nil {
		if errors.Is(err, vmModel.NonUniqueFieldErr) {
			return NonUniqueFieldErr
		}

		return makeError(err)
	}

	err = cs_service.Create(id, params)
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
			if errors.Is(err, vmModel.NonUniqueFieldErr) {
				return NonUniqueFieldErr
			}

			return makeError(err)
		}

		err = cs_service.Update(id, vmUpdate)
		if err != nil {
			return makeError(err)
		}

		err = k8s_service.Repair(id)
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

func UpdateOwnerAuth(id string, params *body.VmUpdateOwner, auth *service.AuthInfo) (*string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update vm owner. details: %w", err)
	}

	vm, err := vmModel.New().GetByID(id)
	if err != nil {
		return nil, makeError(err)
	}

	if vm == nil {
		return nil, VmNotFoundErr
	}

	if vm.OwnerID == params.NewOwnerID {
		return nil, nil
	}

	doTransfer := false

	if auth.IsAdmin {
		doTransfer = true
	} else if auth.UserID == params.NewOwnerID {
		if params.TransferCode == nil || vm.Transfer == nil || vm.Transfer.Code != *params.TransferCode {
			return nil, InvalidTransferCodeErr
		}

		doTransfer = true
	}

	if doTransfer {
		jobID := uuid.New().String()
		err = job_service.Create(jobID, auth.UserID, jobModel.TypeUpdateVmOwner, map[string]interface{}{
			"id":     id,
			"params": *params,
		})

		if err != nil {
			return nil, makeError(err)
		}

		return &jobID, nil
	}

	/// create a transfer notification
	code := createTransferCode()
	err = vmModel.New().UpdateWithParamsByID(id, &vmModel.UpdateParams{
		TransferUserID: &params.NewOwnerID,
		TransferCode:   &code,
	})
	if err != nil {
		return nil, makeError(err)
	}

	err = notification_service.CreateNotification(uuid.NewString(), params.NewOwnerID, &notificationModel.CreateParams{
		Type: notificationModel.TypeVmTransfer,
		Content: map[string]interface{}{
			"id":     vm.ID,
			"name":   vm.Name,
			"userId": params.OldOwnerID,
			"email":  auth.GetEmail(),
			"code":   code,
		},
	})

	if err != nil {
		return nil, makeError(err)
	}

	return nil, nil
}

func UpdateOwner(id string, params *body.VmUpdateOwner) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update vm owner. details: %w", err)
	}

	vm, err := vmModel.New().GetByID(id)
	if err != nil {
		return makeError(err)
	}

	if vm == nil {
		log.Println("vm", id, "not found when updating owner. assuming it was deleted")
		return nil
	}

	emptyString := ""

	err = vmModel.New().UpdateWithParamsByID(id, &vmModel.UpdateParams{
		OwnerID:        &params.NewOwnerID,
		TransferCode:   &emptyString,
		TransferUserID: &emptyString,
	})
	if err != nil {
		return makeError(err)
	}

	err = cs_service.EnsureOwner(id, params.OldOwnerID)
	if err != nil {
		return makeError(err)
	}

	err = k8s_service.EnsureOwner(id, params.OldOwnerID)
	if err != nil {
		return makeError(err)
	}

	log.Println("vm", id, "owner updated to", params.NewOwnerID)
	return nil
}

func ClearUpdateOwner(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to clear vm owner update. details: %w", err)
	}

	deployment, err := vmModel.New().GetByID(id)
	if err != nil {
		return makeError(err)
	}

	if deployment == nil {
		return VmNotFoundErr
	}

	if deployment.Transfer == nil {
		return nil
	}

	emptyString := ""
	err = vmModel.New().UpdateWithParamsByID(id, &vmModel.UpdateParams{
		TransferUserID: &emptyString,
		TransferCode:   &emptyString,
	})
	if err != nil {
		return makeError(err)
	}

	// TODO: delete notification?

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

	err = cs_service.Delete(vm.ID)
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

	err = cs_service.Repair(id)
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

func GetByTransferCode(code, userID string) (*vmModel.VM, error) {
	return vmModel.New().GetByTransferCode(code, userID)
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

func NameAvailable(name string) (bool, error) {
	exists, err := vmModel.New().ExistsByName(name)
	if err != nil {
		return false, err
	}

	return !exists, nil
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

		err := cs_service.DoCommand(csID, gpuID, command, vm.Zone)
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
	case vmModel.ActivityUpdating:
		if vm.DoingOneOfActivities([]string{
			vmModel.ActivityBeingCreated,
			vmModel.ActivityBeingDeleted,
			vmModel.ActivityAttachingGPU,
			vmModel.ActivityDetachingGPU,
		}) {
			return false, "Resource should not be in creation, deletion, and should not be attaching or detaching a GPU", nil
		}
		return true, "", nil
	case vmModel.ActivityAttachingGPU:
		if vm.DoingOneOfActivities([]string{
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
		if vm.DoingOneOfActivities([]string{
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
		if vm.DoingOneOfActivities([]string{
			vmModel.ActivityBeingCreated,
			vmModel.ActivityBeingDeleted,
			vmModel.ActivityAttachingGPU,
			vmModel.ActivityDetachingGPU,
		}) {
			return false, "Resource should not be in creation or deletion, and should not be attaching or detaching a GPU", nil
		}
		return true, "", nil
	case vmModel.ActivityCreatingSnapshot:
		if vm.DoingOneOfActivities([]string{
			vmModel.ActivityBeingCreated,
			vmModel.ActivityBeingDeleted,
			vmModel.ActivityAttachingGPU,
			vmModel.ActivityDetachingGPU,
		}) {
			return false, "Resource should not be in creation or deletion, and should not be attaching or detaching a GPU", nil
		}
		return true, "", nil
	case vmModel.ActivityApplyingSnapshot:
		if vm.DoingOneOfActivities([]string{
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

func CheckQuotaCreate(userID string, quota *roleModel.Quotas, auth *service.AuthInfo, createParams body.VmCreate) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check quota. details: %w", err)
	}

	if auth.IsAdmin {
		return nil
	}

	usage, err := GetUsageByUserID(userID)
	if err != nil {
		return makeError(err)
	}

	totalCpuCores := usage.CpuCores + createParams.CpuCores
	totalRam := usage.RAM + createParams.RAM
	totalDiskSize := usage.DiskSize + createParams.DiskSize

	if totalCpuCores > quota.CpuCores {
		return service.NewQuotaExceededError(fmt.Sprintf("CPU cores quota exceeded. Current: %d, Quota: %d", totalCpuCores, quota.CpuCores))
	}

	if totalRam > quota.RAM {
		return service.NewQuotaExceededError(fmt.Sprintf("RAM quota exceeded. Current: %d, Quota: %d", totalRam, quota.RAM))
	}

	if totalDiskSize > quota.DiskSize {
		return service.NewQuotaExceededError(fmt.Sprintf("Disk size quota exceeded. Current: %d, Quota: %d", totalDiskSize, quota.DiskSize))
	}

	return nil
}

func CheckQuotaUpdate(userID, vmID string, quota *roleModel.Quotas, auth *service.AuthInfo, updateParams body.VmUpdate) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check quota. details: %w", err)
	}

	if auth.IsAdmin {
		return nil
	}

	if updateParams.CpuCores == nil && updateParams.RAM == nil {
		return nil
	}

	usage, err := GetUsageByUserID(userID)
	if err != nil {
		return makeError(err)
	}

	if usage == nil {
		return makeError(fmt.Errorf("usage not found"))
	}

	vm, err := vmModel.New().GetByID(vmID)
	if err != nil {
		return makeError(err)
	}

	if vm == nil {
		return makeError(fmt.Errorf("vm not found"))
	}

	if updateParams.CpuCores != nil {
		totalCpuCores := usage.CpuCores
		if updateParams.CpuCores != nil {
			totalCpuCores += *updateParams.CpuCores - vm.Specs.CpuCores
		}

		if totalCpuCores > quota.CpuCores {
			return service.NewQuotaExceededError(fmt.Sprintf("CPU cores quota exceeded. Current: %d, Quota: %d", totalCpuCores, quota.CpuCores))
		}
	}

	if updateParams.RAM != nil {
		totalRam := usage.RAM
		if updateParams.RAM != nil {
			totalRam += *updateParams.RAM - vm.Specs.RAM
		}

		if totalRam > quota.RAM {
			return service.NewQuotaExceededError(fmt.Sprintf("RAM quota exceeded. Current: %d, Quota: %d", totalRam, quota.RAM))
		}
	}

	return nil
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

func createTransferCode() string {
	return utils.HashString(uuid.NewString())
}
