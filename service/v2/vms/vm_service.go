package vms

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go-deploy/models/dto/v2/body"
	"go-deploy/models/sys/gpu"
	jobModels "go-deploy/models/sys/job"
	notificationModels "go-deploy/models/sys/notification"
	roleModels "go-deploy/models/sys/role"
	teamModels "go-deploy/models/sys/team"
	vmModels "go-deploy/models/sys/vm"
	"go-deploy/models/sys/vm_port"
	"go-deploy/models/versions"
	"go-deploy/pkg/config"
	sErrors "go-deploy/service/errors"
	serviceUtils "go-deploy/service/utils"
	"go-deploy/service/v2/vms/opts"
	"go-deploy/utils"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"sort"
)

// Get gets an existing deployment.
//
// It can be fetched in multiple ways including ID, name, transfer code, and Harbor webhook.
// It supports service.AuthInfo, and will restrict the result to ensure the user has access to the resource.
func (c *Client) Get(id string, opts ...opts.GetOpts) (*vmModels.VM, error) {
	o := serviceUtils.GetFirstOrDefault(opts)

	vmc := vmModels.New(versions.V2)

	if o.TransferCode != nil {
		return vmc.WithTransferCode(*o.TransferCode).Get()
	}

	var effectiveUserID string
	if c.V2.Auth() != nil && !c.V2.Auth().IsAdmin {
		effectiveUserID = c.V2.Auth().UserID
	}

	var teamCheck bool
	if !o.Shared {
		teamCheck = false
	} else if !c.V2.HasAuth() || c.V2.Auth().IsAdmin {
		teamCheck = true
	} else {
		var err error
		teamCheck, err = teamModels.New().WithUserID(c.V2.Auth().UserID).WithResourceID(id).ExistsAny()
		if err != nil {
			return nil, err
		}
	}

	if !teamCheck && effectiveUserID != "" {
		vmc.RestrictToOwner(effectiveUserID)
	}

	return c.VM(id, vmc)
}

// List lists VMs.
//
// It supports service.AuthInfo, and will restrict the result to ensure the user has access to the resource.
func (c *Client) List(opts ...opts.ListOpts) ([]vmModels.VM, error) {
	o := serviceUtils.GetFirstOrDefault(opts)

	vmc := vmModels.New(versions.V2)

	if o.Pagination != nil {
		vmc.WithPagination(o.Pagination.Page, o.Pagination.PageSize)
	}

	var effectiveUserID string
	if o.UserID != nil {
		// Specific user's VMs are requested
		if !c.V2.HasAuth() || c.V2.Auth().UserID == *o.UserID || c.V2.Auth().IsAdmin {
			effectiveUserID = *o.UserID
		} else {
			// User cannot access the other user's resources
			effectiveUserID = c.V2.Auth().UserID
		}
	} else {
		// All VMs are requested
		if c.V2.Auth() != nil && !c.V2.Auth().IsAdmin {
			effectiveUserID = c.V2.Auth().UserID
		}
	}

	if effectiveUserID != "" {
		vmc.RestrictToOwner(effectiveUserID)
	}

	resources, err := c.VMs(vmc)
	if err != nil {
		return nil, err
	}

	// Can only view shared if we are listing resources for a specific user
	if o.Shared && effectiveUserID != "" {
		skipIDs := make([]string, len(resources))
		for i, resource := range resources {
			skipIDs[i] = resource.ID
		}

		teamClient := teamModels.New().WithUserID(effectiveUserID)
		if o.Pagination != nil {
			teamClient.WithPagination(o.Pagination.Page, o.Pagination.PageSize)
		}

		teams, err := teamClient.List()
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
					resources = append(resources, *vm)
				}
			}
		}

		sort.Slice(resources, func(i, j int) bool {
			return resources[i].CreatedAt.After(resources[j].CreatedAt)
		})

		// Since we fetched from two collections, we need to do pagination manually
		if o.Pagination != nil {
			resources = utils.GetPage(resources, o.Pagination.PageSize, o.Pagination.Page)
		}

	} else {
		// Sort by createdAt
		sort.Slice(resources, func(i, j int) bool {
			return resources[i].CreatedAt.After(resources[j].CreatedAt)
		})
	}

	return resources, nil
}

// Create creates a new VM.
//
// It returns an error if the VM already exists (name clash).
func (c *Client) Create(id, ownerID string, dtoVmCreate *body.VmCreate) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create vm. details: %w", err)
	}

	// Right now need to make sure the zone deployed to has KubeVirt installed, so it is hardcoded
	zone := "se-flem-2"
	params := vmModels.CreateParams{}.FromDTOv2(dtoVmCreate, &zone)

	_, err := vmModels.New(versions.V2).Create(id, ownerID, config.Config.Manager, &params)
	if err != nil {
		if errors.Is(err, vmModels.NonUniqueFieldErr) {
			return sErrors.NonUniqueFieldErr
		}

		return makeError(err)
	}

	err = c.K8s().Create(id, &params)
	if err != nil {
		return makeError(err)
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

	vmUpdate := vmModels.UpdateParams{}.FromDTOv2(dtoVmUpdate)

	// We don't allow both applying a snapshot and updating the VM at the same time.
	// So, if a snapshot ID is specified, apply it
	//if vmUpdate.SnapshotID != nil {
	//	err := c.ApplySnapshot(id, *vmUpdate.SnapshotID)
	//	if err != nil {
	//		return makeError(err)
	//	}
	//
	//	return nil
	//}

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

	err := vmModels.New(versions.V2).UpdateWithParams(id, &vmUpdate)
	if err != nil {
		if errors.Is(err, vmModels.NonUniqueFieldErr) {
			return sErrors.NonUniqueFieldErr
		}

		return makeError(err)
	}

	_, err = c.Refresh(id)
	if err != nil {
		return makeError(err)
	}

	err = c.K8s().Repair(id)
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

	nmc := notificationModels.New().FilterContent("id", id)
	err = nmc.Delete()
	if err != nil {
		return makeError(err)
	}

	err = c.K8s().Delete(id)
	if err != nil {
		return makeError(err)
	}

	err = gpu.New().Detach(vm.ID)
	if err != nil {
		return makeError(err)
	}

	err = vm_port.New().ReleaseAll(vm.ID)
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

	err = c.K8s().Repair(id)
	if err != nil {
		return makeError(err)
	}

	log.Println("repaired vm", id)
	return nil
}

// DoAction performs an action on the VM.
func (c *Client) DoAction(id string, dtoVmAction *body.VmAction) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to perform action on vm %s. details: %w", id, err)
	}

	params := vmModels.ActionParams{}.FromDTOv2(dtoVmAction)

	vm, err := c.VM(id, nil)
	if err != nil {
		return makeError(err)
	}

	if vm == nil {
		log.Println("vm", id, "not found when performing action. assuming it was deleted")
		return nil
	}

	err = c.K8s().DoAction(id, &params)
	if err != nil {
		return makeError(err)
	}

	log.Println("performed action", params.Action, "on vm", id)
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

	if !c.V2.HasAuth() || c.V2.Auth().IsAdmin {
		transferDirectly = true
	} else if c.V2.Auth().UserID == params.NewOwnerID {
		if params.TransferCode == nil || vm.Transfer == nil || vm.Transfer.Code != *params.TransferCode {
			return nil, sErrors.InvalidTransferCodeErr
		}

		transferDirectly = true
	}

	if transferDirectly {
		jobID := uuid.New().String()
		err = c.V1.Jobs().Create(jobID, c.V2.Auth().UserID, jobModels.TypeUpdateVmOwner, versions.V2, map[string]interface{}{
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
	err = vmModels.New(versions.V2).UpdateWithParams(id, &vmModels.UpdateParams{
		TransferUserID: &params.NewOwnerID,
		TransferCode:   &code,
	})
	if err != nil {
		return nil, makeError(err)
	}

	_, err = c.V1.Notifications().Create(uuid.NewString(), params.NewOwnerID, &notificationModels.CreateParams{
		Type: notificationModels.TypeVmTransfer,
		Content: map[string]interface{}{
			"id":     vm.ID,
			"name":   vm.Name,
			"userId": params.OldOwnerID,
			"email":  c.V2.Auth().GetEmail(),
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

	err = vmModels.New(versions.V2).UpdateWithParams(id, &vmModels.UpdateParams{
		OwnerID:        &params.NewOwnerID,
		TransferCode:   &emptyString,
		TransferUserID: &emptyString,
	})
	if err != nil {
		return makeError(err)
	}

	err = gpu.New().WithVM(id).UpdateWithBSON(bson.D{{"lease.user", params.NewOwnerID}})
	if err != nil {
		return makeError(err)
	}

	err = c.K8s().EnsureOwner(id, params.OldOwnerID)
	if err != nil {
		return makeError(err)
	}

	nmc := notificationModels.New().FilterContent("id", id).WithType(notificationModels.TypeVmTransfer)
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

	deployment, err := vmModels.New(versions.V2).GetByID(id)
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
	err = vmModels.New(versions.V2).UpdateWithParams(id, &vmModels.UpdateParams{
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

	return nil, makeError(errors.New("not implemented"))
}

// IsBeingDeleted checks if the VM is being deleted.
//
// This returns true while the VM is being deleted, and after it has been deleted.
func (c *Client) IsBeingDeleted(id string) (bool, error) {
	vm, err := c.VM(id, nil)
	if err != nil {
		return false, err
	}

	if vm == nil {
		return true, nil
	}

	return vm.BeingDeleted(), nil
}

// NameAvailable checks if the given name is available.
func (c *Client) NameAvailable(name string) (bool, error) {
	exists, err := vmModels.New().ExistsByName(name)
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

	exists, err := vmModels.New().WithCustomFilter(filter).ExistsAny()
	if err != nil {
		return false, err
	}

	return !exists, nil
}

// SshConnectionString gets the SSH connection string for the VM.
//
// It returns nil if the VM is not found.
func (c *Client) SshConnectionString(id string) (*string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get SSH connection string for vm %s. details: %w", id, err)
	}

	vm, err := c.VM(id, nil)
	if err != nil {
		return nil, makeError(err)
	}

	if vm == nil {
		return nil, nil
	}

	zone := config.Config.Deployment.GetZone(vm.Zone)
	if zone == nil {
		return nil, makeError(sErrors.ZoneNotFoundErr)
	}

	var sshConnectionString *string
	if service := vm.Subsystems.K8s.GetService(fmt.Sprintf("%s-priv-22-prot-tcp", vm.Name)); service != nil {
		for _, port := range service.Ports {
			if port.TargetPort == 22 {
				sshConnectionString = utils.StrPtr(fmt.Sprintf("ssh root@%s -p %d", zone.ParentDomainVM, port.Port))
			}
		}
	}

	return sshConnectionString, nil
}

// CheckQuota checks if the user has enough quota to create or update a deployment.
//
// Make sure to specify either opts.Create or opts.Update in the options (opts.Create takes priority).
// When checking quota for opts.Create and opts.CreateSnapshot, id is not used.
//
// It returns an error if the user does not have enough quotas.
func (c *Client) CheckQuota(id, userID string, quota *roleModels.Quotas, opts ...opts.QuotaOpts) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check quota for user %s. details: %w", userID, err)
	}

	if !c.V2.HasAuth() || c.V2.Auth().IsAdmin {
		return nil
	}

	o := serviceUtils.GetFirstOrDefault(opts)

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

		vm, err := vmModels.New(versions.V2).GetByID(id)
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
func (c *Client) GetUsage(userID string) (*vmModels.Usage, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get usage for user %s. details: %w", userID, err)
	}

	usage := &vmModels.Usage{}

	currentVms, err := vmModels.New(versions.V2).RestrictToOwner(userID).List()
	if err != nil {
		return nil, makeError(err)
	}

	for _, vm := range currentVms {
		specs := vm.Specs

		usage.CpuCores += specs.CpuCores
		usage.RAM += specs.RAM
		usage.DiskSize += specs.DiskSize

		// TODO: Add snapshot usage
	}

	return usage, nil
}

// GetHost gets the host for the VM.
//
// It returns an error if the VM is not found.
func (c *Client) GetHost(vmID string) (*vmModels.Host, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get host for vm %s. details: %w", vmID, err)
	}

	return nil, makeError(errors.New("not implemented"))
}

// createTransferCode creates a transfer code.
func createTransferCode() string {
	return utils.HashStringAlphanumeric(uuid.NewString())
}

// portName returns the name of a port used as a key in the port map in the database.
func portName(privatePort int, protocol string) string {
	return fmt.Sprintf("priv-%d-prot-%s", privatePort, protocol)
}
