package vms

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go-deploy/dto/v1/body"
	"go-deploy/models/model"
	"go-deploy/models/version"
	"go-deploy/pkg/config"
	"go-deploy/pkg/db/resources/gpu_repo"
	"go-deploy/pkg/db/resources/notification_repo"
	"go-deploy/pkg/db/resources/team_repo"
	"go-deploy/pkg/db/resources/vm_port_repo"
	"go-deploy/pkg/db/resources/vm_repo"
	sErrors "go-deploy/service/errors"
	sUtils "go-deploy/service/utils"
	"go-deploy/service/v1/vms/cs_service"
	"go-deploy/service/v1/vms/k8s_service"
	"go-deploy/service/v1/vms/opts"
	"go-deploy/utils"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"sort"
	"strings"
)

// Get gets an existing deployment.
//
// It can be fetched in multiple ways including ID, name, transfer code, and Harbor webhook.
// It supports service.AuthInfo, and will restrict the result to ensure the user has access to the model.
func (c *Client) Get(id string, opts ...opts.GetOpts) (*model.VM, error) {
	o := sUtils.GetFirstOrDefault(opts)

	vmc := vm_repo.New(version.V1)

	if o.TransferCode != nil {
		return vmc.WithTransferCode(*o.TransferCode).Get()
	}

	var effectiveUserID string
	if c.V1.Auth() != nil && !c.V1.Auth().IsAdmin {
		effectiveUserID = c.V1.Auth().UserID
	}

	var teamCheck bool
	if !o.Shared {
		teamCheck = false
	} else if !c.V1.HasAuth() || c.V1.Auth().IsAdmin {
		teamCheck = true
	} else {
		var err error
		teamCheck, err = team_repo.New().WithUserID(c.V1.Auth().UserID).WithResourceID(id).ExistsAny()
		if err != nil {
			return nil, err
		}
	}

	if !teamCheck && effectiveUserID != "" {
		vmc.WithOwner(effectiveUserID)
	}

	return c.VM(id, vmc)
}

// List lists VMs.
//
// It supports service.AuthInfo, and will restrict the result to ensure the user has access to the model.
func (c *Client) List(opts ...opts.ListOpts) ([]model.VM, error) {
	o := sUtils.GetFirstOrDefault(opts)

	vmc := vm_repo.New(version.V1)

	if o.Pagination != nil {
		vmc.WithPagination(o.Pagination.Page, o.Pagination.PageSize)
	}

	var effectiveUserID string
	if o.UserID != nil {
		// Specific user's VMs are requested
		if !c.V1.HasAuth() || c.V1.Auth().UserID == *o.UserID || c.V1.Auth().IsAdmin {
			effectiveUserID = *o.UserID
		} else {
			// User cannot access the other user's resources
			effectiveUserID = c.V1.Auth().UserID
		}
	} else {
		// All VMs are requested
		if c.V1.Auth() != nil && !c.V1.Auth().IsAdmin {
			effectiveUserID = c.V1.Auth().UserID
		}
	}

	if effectiveUserID != "" {
		vmc.WithOwner(effectiveUserID)
	}

	vms, err := c.VMs(vmc)
	if err != nil {
		return nil, err
	}

	// Can only view shared if we are listing resources for a specific user
	if o.Shared && effectiveUserID != "" {
		skipIDs := make([]string, len(vms))
		for i, resource := range vms {
			skipIDs[i] = resource.ID
		}

		teamClient := team_repo.New().WithUserID(effectiveUserID)
		if o.Pagination != nil {
			teamClient.WithPagination(o.Pagination.Page, o.Pagination.PageSize)
		}

		teams, err := teamClient.List()
		if err != nil {
			return nil, err
		}

		for _, team := range teams {
			for _, resource := range team.GetResourceMap() {
				if resource.Type != model.TeamResourceVM {
					continue
				}

				// skip existing non-shared resources
				skip := false
				for _, skipID := range skipIDs {
					if resource.ID == skipID {
						skip = true
						break
					}
				}
				if skip {
					continue
				}

				vm, err := c.VM(resource.ID, nil)
				if err != nil {
					return nil, err
				}

				if vm != nil {
					vms = append(vms, *vm)
				}
			}
		}

		sort.Slice(vms, func(i, j int) bool {
			return vms[i].CreatedAt.After(vms[j].CreatedAt)
		})

		// Since we fetched from two collections, we need to do pagination manually
		if o.Pagination != nil {
			vms = utils.GetPage(vms, o.Pagination.PageSize, o.Pagination.Page)
		}

	} else {
		// Sort by createdAt
		sort.Slice(vms, func(i, j int) bool {
			return vms[i].CreatedAt.After(vms[j].CreatedAt)
		})
	}

	return vms, nil
}

// Create creates a new VM.
//
// It returns an error if the VM already exists (name clash).
func (c *Client) Create(id, ownerID string, dtoVmCreate *body.VmCreate) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create vm. details: %w", err)
	}

	// Temporary hard-coded fallback
	fallback := "se-flem"
	deploymentZone := "se-flem"

	params := model.VmCreateParams{}.FromDTOv1(dtoVmCreate, &fallback, &deploymentZone)

	_, err := vm_repo.New(version.V1).Create(id, ownerID, config.Config.Manager, &params)
	if err != nil {
		if errors.Is(err, vm_repo.NonUniqueFieldErr) {
			return sErrors.NonUniqueFieldErr
		}

		return makeError(err)
	}

	err = cs_service.New(c.Cache).Create(id, &params)
	if err != nil {
		return makeError(err)
	}

	if len(params.PortMap) > 1 {
		err = k8s_service.New(c.Cache).Create(id, &params)
		if err != nil {
			return makeError(err)
		}
	} else {
		log.Println("skipping k8s setup for vm", id, "since it has no ports")
	}

	return nil
}

// Update updates an existing VM.
//
// It returns an error if the VM is not found.
func (c *Client) Update(id string, dtoVmUpdate *body.VmUpdate) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update vm. details: %w", err)
	}

	vmUpdate := model.VmUpdateParams{}.FromDTOv1(dtoVmUpdate)

	// We don't allow both applying a snapshot and updating the VM at the same time.
	// So, if a snapshot ID is specified, apply it
	if vmUpdate.SnapshotID != nil {
		err := c.ApplySnapshot(id, *vmUpdate.SnapshotID)
		if err != nil {
			return makeError(err)
		}

		return nil
	}

	// Otherwise, update the VM as usual
	if vmUpdate.PortMap != nil {
		// We don't want to give new secrets for the same custom domains
		vm, err := c.VM(id, nil)
		if err != nil {
			return makeError(err)
		}

		// So we find if there are any custom domains that are being updated with the same domain name,
		// and if so, we remove the update from the params
		for name, p1 := range vm.PortMap {
			if p2, ok := (*vmUpdate.PortMap)[name]; ok {
				if p1.HttpProxy != nil && p2.HttpProxy != nil && p1.HttpProxy.CustomDomain != nil && p2.HttpProxy.CustomDomain != nil {
					if p1.HttpProxy.CustomDomain.Domain == *p2.HttpProxy.CustomDomain {
						p2.HttpProxy.CustomDomain = nil
					}
				}
			}
		}
	}

	err := vm_repo.New(version.V1).UpdateWithParams(id, &vmUpdate)
	if err != nil {
		if errors.Is(err, vm_repo.NonUniqueFieldErr) {
			return sErrors.NonUniqueFieldErr
		}

		return makeError(err)
	}

	_, err = c.Refresh(id)
	if err != nil {
		return makeError(err)
	}

	err = cs_service.New(c.Cache).Update(id, &vmUpdate)
	if err != nil {
		return makeError(err)
	}

	err = k8s_service.New(c.Cache).Repair(id)
	if err != nil {
		return makeError(err)
	}

	return nil
}

// Delete deletes an existing VM.
//
// It returns an error if the VM is not found.
func (c *Client) Delete(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete vm. details: %w", err)
	}

	vm, err := c.VM(id, nil)
	if err != nil {
		return makeError(err)
	}

	if vm == nil {
		return sErrors.VmNotFoundErr
	}

	nmc := notification_repo.New().FilterContent("id", id)
	err = nmc.Delete()
	if err != nil {
		return makeError(err)
	}

	err = k8s_service.New(c.Cache).Delete(id)
	if err != nil {
		return makeError(err)
	}

	err = cs_service.New(c.Cache).Delete(vm.ID)
	if err != nil {
		return makeError(err)
	}

	err = gpu_repo.New().Detach(vm.ID)
	if err != nil {
		return makeError(err)
	}

	err = vm_port_repo.New().ReleaseAll(vm.ID)
	if err != nil {
		return makeError(err)
	}

	return nil
}

// Repair repairs an existing deployment.
//
// Trigger repair jobs for every subsystem.
func (c *Client) Repair(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to repair vm %s. details: %w", id, err)
	}

	vm, err := c.Get(id)
	if err != nil {
		return makeError(err)
	}

	if vm == nil {
		log.Println("vm", id, "not found when repairing. assuming it was deleted")
		return nil
	}

	err = cs_service.New(c.Cache).Repair(id)
	if err != nil {
		return makeError(err)
	}

	err = k8s_service.New(c.Cache).Repair(id)
	if err != nil {
		return makeError(err)
	}

	log.Println("repaired vm", id)
	return nil
}

// UpdateOwnerSetup updates the owner of the VM.
//
// This is the first step of the owner update process, where it is decided if a notification should be created,
// or if the transfer should be done immediately.
//
// It returns an error if the VM is not found.
func (c *Client) UpdateOwnerSetup(id string, params *body.VmUpdateOwner) (*string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update vm owner. details: %w", err)
	}

	vm, err := c.VM(id, nil)
	if err != nil {
		return nil, makeError(err)
	}

	if vm == nil {
		return nil, sErrors.VmNotFoundErr
	}

	if vm.OwnerID == params.NewOwnerID {
		return nil, nil
	}

	transferDirectly := false

	if !c.V1.HasAuth() || c.V1.Auth().IsAdmin {
		transferDirectly = true
	} else if c.V1.Auth().UserID == params.NewOwnerID {
		if params.TransferCode == nil || vm.Transfer == nil || vm.Transfer.Code != *params.TransferCode {
			return nil, sErrors.InvalidTransferCodeErr
		}

		transferDirectly = true
	}

	if transferDirectly {
		jobID := uuid.New().String()
		err = c.V1.Jobs().Create(jobID, c.V1.Auth().UserID, model.JobUpdateVmOwner, version.V1, map[string]interface{}{
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
	err = vm_repo.New(version.V1).UpdateWithParams(id, &model.VmUpdateParams{
		TransferUserID: &params.NewOwnerID,
		TransferCode:   &code,
	})
	if err != nil {
		return nil, makeError(err)
	}

	_, err = c.V1.Notifications().Create(uuid.NewString(), params.NewOwnerID, &model.NotificationCreateParams{
		Type: model.NotificationVmTransfer,
		Content: map[string]interface{}{
			"id":     vm.ID,
			"name":   vm.Name,
			"userId": params.OldOwnerID,
			"email":  c.V1.Auth().GetEmail(),
			"code":   code,
		},
	})

	if err != nil {
		return nil, makeError(err)
	}

	return nil, nil
}

// UpdateOwner updates the owner of the VM.
//
// This is the second step of the owner update process, where the transfer is actually done.
//
// It returns an error if the VM is not found.
func (c *Client) UpdateOwner(id string, params *body.VmUpdateOwner) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update vm owner. details: %w", err)
	}

	vm, err := c.VM(id, nil)
	if err != nil {
		return makeError(err)
	}

	if vm == nil {
		return sErrors.VmNotFoundErr
	}

	if vm == nil {
		log.Println("vm", id, "not found when updating owner. assuming it was deleted")
		return nil
	}

	emptyString := ""

	err = vm_repo.New(version.V1).UpdateWithParams(id, &model.VmUpdateParams{
		OwnerID:        &params.NewOwnerID,
		TransferCode:   &emptyString,
		TransferUserID: &emptyString,
	})
	if err != nil {
		return makeError(err)
	}

	err = gpu_repo.New().WithVM(id).UpdateWithBSON(bson.D{{"lease.user", params.NewOwnerID}})
	if err != nil {
		return makeError(err)
	}

	err = cs_service.New(c.Cache).EnsureOwner(id, params.OldOwnerID)
	if err != nil {
		return makeError(err)
	}

	err = k8s_service.New(c.Cache).EnsureOwner(id, params.OldOwnerID)
	if err != nil {
		return makeError(err)
	}

	nmc := notification_repo.New().FilterContent("id", id).WithType(model.NotificationVmTransfer)
	err = nmc.MarkReadAndCompleted()
	if err != nil {
		return makeError(err)
	}

	log.Println("vm", id, "owner updated from", params.OldOwnerID, " to", params.NewOwnerID)
	return nil
}

// ClearUpdateOwner clears the owner update process.
//
// This is intended to be used when the owner update process is canceled.
func (c *Client) ClearUpdateOwner(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to clear vm owner update. details: %w", err)
	}

	deployment, err := vm_repo.New(version.V1).GetByID(id)
	if err != nil {
		return makeError(err)
	}

	if deployment == nil {
		return sErrors.VmNotFoundErr
	}

	if deployment.Transfer == nil {
		return nil
	}

	emptyString := ""
	err = vm_repo.New(version.V1).UpdateWithParams(id, &model.VmUpdateParams{
		TransferUserID: &emptyString,
		TransferCode:   &emptyString,
	})
	if err != nil {
		return makeError(err)
	}

	// TODO: delete notification?

	return nil
}

// GetConnectionString gets the connection string for the VM.
//
// It returns nil if the VM is not found.
func (c *Client) GetConnectionString(id string) (*string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get connection string for vm %s. details: %w", id, err)
	}

	vm, err := c.VM(id, nil)
	if err != nil {
		return nil, makeError(err)
	}

	zone := config.Config.VM.GetZone(vm.Zone)
	if zone == nil {
		return nil, makeError(sErrors.ZoneNotFoundErr)
	}

	domainName := zone.ParentDomain
	rule := vm.Subsystems.CS.GetPortForwardingRule("__ssh")
	if rule == nil {
		return nil, nil
	}

	if domainName == "" || rule.PublicPort == 0 {
		return nil, nil
	}

	connectionString := fmt.Sprintf("ssh root@%s -p %d", domainName, rule.PublicPort)

	return &connectionString, nil
}

// DoCommand executes a command on the VM.
//
// It is purely best-effort, and does not return any errors.
func (c *Client) DoCommand(id, command string) {
	go func() {
		vm, err := c.VM(id, nil)
		if err != nil {
			log.Println("failed to get vm", id, "when executing command", command, "details:", err)
			return
		}

		if vm == nil {
			log.Println("vm", id, "not found when executing command", command, ". assuming it was deleted")
			return
		}

		csID := vm.Subsystems.CS.VM.ID
		if csID == "" {
			log.Println("cs vm not setup when executing command", command, "for vm", id, ". assuming it was deleted")
			return
		}

		gpuID, err := gpu_repo.New().WithVM(vm.ID).GetID()
		if err != nil {
			log.Println("failed to get gpu id for vm", id, "when executing command", command, "details:", err)
			return
		}

		err = cs_service.New(c.Cache).DoCommand(vm.ID, csID, gpuID, command)
		if err != nil {
			utils.PrettyPrintError(err)
			return
		}
	}()
}

// StartActivity starts an activity for the deployment.
//
// It only starts the activity if it is allowed, determined by CanAddActivity.
// It returns a boolean indicating if the activity was started, and a string indicating the reason if it was not.
func (c *Client) StartActivity(id, activity string) error {
	canAdd, reason, err := c.CanAddActivity(id, activity)
	if !canAdd {
		if reason == "Deployment not found" {
			return sErrors.DeploymentNotFoundErr
		}

		return sErrors.NewFailedToStartActivityError(reason)
	}

	err = vm_repo.New(version.V1).AddActivity(id, activity)
	if err != nil {
		return err
	}

	return nil
}

// CanAddActivity checks if the deployment can add an activity.
//
// It returns a boolean indicating if the activity can be added, and a string indicating the reason if it cannot.
func (c *Client) CanAddActivity(vmID, activity string) (bool, string, error) {
	vm, err := c.VM(vmID, nil)
	if err != nil {
		return false, "", err
	}

	if vm == nil {
		return false, "", err
	}

	switch activity {
	case model.ActivityBeingCreated:
		return !vm.BeingDeleted(), "Resource is being deleted", nil
	case model.ActivityBeingDeleted:
		return true, "", nil
	case model.ActivityUpdating:
		if vm.DoingOneOfActivities([]string{
			model.ActivityBeingCreated,
			model.ActivityBeingDeleted,
			model.ActivityAttachingGPU,
			model.ActivityDetachingGPU,
		}) {
			return false, "Resource should not be in creation, deletion, and should not be attaching or detaching a GPU", nil
		}
		return true, "", nil
	case model.ActivityAttachingGPU:
		if vm.DoingOneOfActivities([]string{
			model.ActivityBeingCreated,
			model.ActivityBeingDeleted,
			model.ActivityAttachingGPU,
			model.ActivityDetachingGPU,
		}) {
			return false, "Resource should not be in creation or deletion, and should not be attaching or detaching a GPU", nil
		}
		return true, "", nil
	case model.ActivityDetachingGPU:
		if vm.DoingOneOfActivities([]string{
			model.ActivityBeingCreated,
			model.ActivityBeingDeleted,
			model.ActivityAttachingGPU,
		}) {
			return false, "Resource should not be in creation or deletion, and should not be attaching a GPU", nil
		}
		return true, "", nil
	case model.ActivityRepairing:
		if vm.DoingOneOfActivities([]string{
			model.ActivityBeingCreated,
			model.ActivityBeingDeleted,
			model.ActivityAttachingGPU,
			model.ActivityDetachingGPU,
		}) {
			return false, "Resource should not be in creation or deletion, and should not be attaching or detaching a GPU", nil
		}
		return true, "", nil
	}

	return false, "", fmt.Errorf("unknown activity %s", activity)
}

// NameAvailable checks if the given name is available.
func (c *Client) NameAvailable(name string) (bool, error) {
	exists, err := vm_repo.New().ExistsByName(name)
	if err != nil {
		return false, err
	}

	return !exists, nil
}

// HttpProxyNameAvailable checks if the given name is available for an HTTP proxy.
func (c *Client) HttpProxyNameAvailable(id, name string) (bool, error) {
	filter := bson.D{
		{"id", bson.D{{"$ne", id}}},
		{"portMap.httpProxy.name", name},
	}

	exists, err := vm_repo.New().WithCustomFilter(filter).ExistsAny()
	if err != nil {
		return false, err
	}

	return !exists, nil
}

// CheckQuota checks if the user has enough quota to create or update a deployment.
//
// Make sure to specify either opts.Create or opts.Update in the options (opts.Create takes priority).
// When checking quota for opts.Create and opts.CreateSnapshot, id is not used.
//
// It returns an error if the user does not have enough quotas.
func (c *Client) CheckQuota(id, userID string, quota *model.Quotas, opts ...opts.QuotaOpts) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check quota for user %s. details: %w", userID, err)
	}

	if !c.V1.HasAuth() || c.V1.Auth().IsAdmin {
		return nil
	}

	o := sUtils.GetFirstOrDefault(opts)

	usage, err := c.GetUsage(userID)
	if err != nil {
		return makeError(err)
	}

	if usage == nil {
		return makeError(fmt.Errorf("failed to get usage for user %s", userID))
	}

	if o.Create != nil {
		totalCpuCores := usage.CpuCores + o.Create.CpuCores
		totalRam := usage.RAM + o.Create.RAM
		totalDiskSize := usage.DiskSize + o.Create.DiskSize

		if totalCpuCores > quota.CpuCores {
			return sErrors.NewQuotaExceededError(fmt.Sprintf("CPU cores quota exceeded. Current: %d, Quota: %d", totalCpuCores, quota.CpuCores))
		}

		if totalRam > quota.RAM {
			return sErrors.NewQuotaExceededError(fmt.Sprintf("RAM quota exceeded. Current: %d, Quota: %d", totalRam, quota.RAM))
		}

		if totalDiskSize > quota.DiskSize {
			return sErrors.NewQuotaExceededError(fmt.Sprintf("Disk size quota exceeded. Current: %d, Quota: %d", totalDiskSize, quota.DiskSize))
		}
	} else if o.Update != nil {
		if o.Update.CpuCores == nil && o.Update.RAM == nil {
			return nil
		}

		vm, err := vm_repo.New(version.V1).GetByID(id)
		if err != nil {
			return makeError(err)
		}

		if vm == nil {
			return makeError(sErrors.VmNotFoundErr)
		}

		if o.Update.CpuCores != nil {
			totalCpuCores := usage.CpuCores
			if o.Update.CpuCores != nil {
				totalCpuCores += *o.Update.CpuCores - vm.Specs.CpuCores
			}

			if totalCpuCores > quota.CpuCores {
				return sErrors.NewQuotaExceededError(fmt.Sprintf("CPU cores quota exceeded. Current: %d, Quota: %d", totalCpuCores, quota.CpuCores))
			}
		}

		if o.Update.RAM != nil {
			totalRam := usage.RAM
			if o.Update.RAM != nil {
				totalRam += *o.Update.RAM - vm.Specs.RAM
			}

			if totalRam > quota.RAM {
				return sErrors.NewQuotaExceededError(fmt.Sprintf("RAM quota exceeded. Current: %d, Quota: %d", totalRam, quota.RAM))
			}
		}
	} else if o.CreateSnapshot != nil {
		if usage.Snapshots >= quota.Snapshots {
			return sErrors.NewQuotaExceededError(fmt.Sprintf("Snapshot count quota exceeded. Current: %d, Quota: %d", usage.Snapshots, quota.Snapshots))
		}
	}

	return nil
}

// GetUsage gets the usage for the user.
//
// If user does not exist, or user does not have any VMs, it returns an empty usage.
func (c *Client) GetUsage(userID string) (*model.VmUsage, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get usage for user %s. details: %w", userID, err)
	}

	usage := &model.VmUsage{}

	currentVms, err := vm_repo.New(version.V1).WithOwner(userID).List()
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

// GetExternalPortMapper gets the external port mapper for the VM.
//
// If the VM is not found, it returns nil.
func (c *Client) GetExternalPortMapper(vmID string) (map[string]int, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get external port mapper. details: %w", err)
	}

	vm, err := c.VM(vmID, nil)
	if err != nil {
		return nil, makeError(err)
	}

	if vm == nil {
		log.Println("vm", vmID, "not found when detaching getting external port mapper. assuming it was deleted")
		return nil, nil
	}

	mapper := make(map[string]int)
	for name, port := range vm.Subsystems.CS.PortForwardingRuleMap {
		mapper[name] = port.PublicPort
	}

	return mapper, nil
}

// GetHost gets the host for the VM.
//
// It returns an error if the VM is not found.
func (c *Client) GetHost(vmID string) (*model.Host, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get host for vm %s. details: %w", vmID, err)
	}

	// 1. Try to get the host from the VM, this only works if the VM is running
	vm, err := c.VM(vmID, nil)
	if err != nil {
		return nil, makeError(err)
	}

	if vm == nil {
		log.Println("vm", vmID, "not found when getting host. assuming it was deleted")
		return nil, nil
	}

	zone := config.Config.VM.GetZone(vm.Zone)
	if zone == nil {
		return nil, makeError(sErrors.ZoneNotFoundErr)
	}

	cc := cs_service.New(c.Cache)

	host, err := cc.GetHostByVM(vmID)
	if err != nil {
		return nil, makeError(err)
	}

	if host != nil {
		return &model.Host{
			ID:   host.ID,
			Name: host.Name,
		}, nil
	}

	// 2. Try to get the host by the GPU, since the VM is required to start on the host with the GPU
	gpuID, err := gpu_repo.New().WithVM(vmID).GetID()
	if err != nil {
		return nil, makeError(err)
	}

	if gpuID != nil {
		hostName, err := cc.GetRequiredHost(*gpuID)
		if err != nil {
			return nil, makeError(err)
		}

		if hostName != nil {
			host, err = cc.GetHostByName(*hostName, zone)
			if err != nil {
				return nil, makeError(err)
			}

			if host != nil {
				return &model.Host{
					ID:   host.ID,
					Name: host.Name,
				}, nil
			}
		}
	}

	// The host was not found using any method
	return nil, nil
}

// GetCloudStackHostCapabilities gets the capabilities of the host, total and used.
func (c *Client) GetCloudStackHostCapabilities(hostName string, zoneName string) (*model.CloudStackHostCapabilities, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get host capabilities. details: %w", err)
	}

	cc := cs_service.New(nil)

	zone := config.Config.VM.GetZone(zoneName)
	if zone == nil {
		return nil, makeError(sErrors.ZoneNotFoundErr)
	}

	host, err := cc.GetHostByName(hostName, zone)
	if err != nil {
		return nil, makeError(err)
	}

	if host == nil {
		return nil, nil
	}

	configuration, err := cc.GetConfiguration(zone)
	if err != nil {
		return nil, makeError(err)
	}

	return &model.CloudStackHostCapabilities{
		CpuCoresTotal: host.CpuCoresTotal * configuration.OverProvisioningFactor,
		CpuCoresUsed:  host.CpuCoresUsed,
		RamTotal:      host.RamTotal,
		RamUsed:       host.RamUsed,
		RamAllocated:  host.RamAllocated,
	}, nil
}

func createTransferCode() string {
	return utils.HashStringAlphanumeric(uuid.NewString())
}
