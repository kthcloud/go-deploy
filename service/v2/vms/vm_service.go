package vms

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go-deploy/dto/v2/body"
	configModels "go-deploy/models/config"
	"go-deploy/models/model"
	"go-deploy/models/version"
	"go-deploy/pkg/config"
	"go-deploy/pkg/db/resources/gpu_lease_repo"
	"go-deploy/pkg/db/resources/gpu_repo"
	"go-deploy/pkg/db/resources/notification_repo"
	"go-deploy/pkg/db/resources/resource_migration_repo"
	"go-deploy/pkg/db/resources/team_repo"
	"go-deploy/pkg/db/resources/vm_port_repo"
	"go-deploy/pkg/db/resources/vm_repo"
	"go-deploy/pkg/log"
	sErrors "go-deploy/service/errors"
	serviceUtils "go-deploy/service/utils"
	"go-deploy/service/v2/vms/opts"
	"go-deploy/utils"
	"go.mongodb.org/mongo-driver/bson"
	"sort"
	"strings"
)

// Get gets an existing deployment.
//
// It can be fetched in multiple ways including ID, name, transfer code, and Harbor webhook.
// It supports service.AuthInfo, and will restrict the result to ensure the user has access to the model.
func (c *Client) Get(id string, opts ...opts.GetOpts) (*model.VM, error) {
	o := serviceUtils.GetFirstOrDefault(opts)

	vmc := vm_repo.New(version.V2)

	if o.MigrationCode != nil {
		rmc := resource_migration_repo.New().
			WithType(model.ResourceMigrationTypeUpdateOwner).
			WithResourceType(model.ResourceMigrationResourceTypeDeployment).
			WithTransferCode(*o.MigrationCode)

		migration, err := rmc.Get()
		if err != nil {
			return nil, err
		}

		if migration == nil {
			return nil, nil
		}

		return c.VM(migration.ResourceID, vmc)
	}

	var effectiveUserID string
	if c.V2.Auth() != nil && !c.V2.Auth().User.IsAdmin {
		effectiveUserID = c.V2.Auth().User.ID
	}

	var teamCheck bool
	if !o.Shared {
		teamCheck = false
	} else if !c.V2.HasAuth() || c.V2.Auth().User.IsAdmin {
		teamCheck = true
	} else {
		var err error
		teamCheck, err = team_repo.New().WithUserID(c.V2.Auth().User.ID).WithResourceID(id).ExistsAny()
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
	o := serviceUtils.GetFirstOrDefault(opts)

	vmc := vm_repo.New(version.V2)

	if o.Pagination != nil {
		vmc.WithPagination(o.Pagination.Page, o.Pagination.PageSize)
	}

	var effectiveUserID string
	if o.UserID != nil {
		// Specific user's VMs are requested
		if !c.V2.HasAuth() || c.V2.Auth().User.ID == *o.UserID || c.V2.Auth().User.IsAdmin {
			effectiveUserID = *o.UserID
		} else {
			// User cannot access the other user's resources
			effectiveUserID = c.V2.Auth().User.ID
		}
	} else {
		// All VMs are requested
		if c.V2.Auth() != nil && !c.V2.Auth().User.IsAdmin {
			effectiveUserID = c.V2.Auth().User.ID
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
		for i, vm := range vms {
			skipIDs[i] = vm.ID
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

	fallbackZone := config.Config.VM.DefaultZone
	params := model.VmCreateParams{}.FromDTOv2(dtoVmCreate, &fallbackZone)

	if !c.V1.Zones().HasCapability(params.Zone, configModels.ZoneCapabilityVM) {
		return sErrors.NewZoneCapabilityMissingError(fallbackZone, configModels.ZoneCapabilityVM)
	}

	_, err := vm_repo.New(version.V2).Create(id, ownerID, &params)
	if err != nil {
		if errors.Is(err, vm_repo.NonUniqueFieldErr) {
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

	vmUpdate := model.VmUpdateParams{}.FromDTOv2(dtoVmUpdate)

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

	err := vm_repo.New(version.V2).UpdateWithParams(id, &vmUpdate)
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

	nmc := notification_repo.New().FilterContent("id", id)
	err = nmc.Delete()
	if err != nil {
		return makeError(err)
	}

	err = c.V1.Teams().CleanResource(id)
	if err != nil {
		return makeError(err)
	}

	err = c.K8s().Delete(id)
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

	err = gpu_lease_repo.New().WithVmID(id).Release()
	if err != nil {
		return makeError(err)
	}

	return nil
}

// IsAccessible checks if the VM is accessible by the caller.
// This is useful when providing auth info and check if it's enough to access a VM.
// It is lightweight since it does not require the VM to be fetched.
//
// It returns an error if the VM is not found.
func (c *Client) IsAccessible(id string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if user can access vm %s. details: %w", id, err)
	}

	exists, err := vm_repo.New(version.V2).ExistsByID(id)
	if err != nil {
		return false, makeError(err)
	}

	if !exists {
		return false, makeError(sErrors.VmNotFoundErr)
	}

	if c.V2.HasAuth() {
		// 1. User has access through being an admin
		if c.V2.Auth().User.IsAdmin {
			return true, nil
		}

		// 2. User has access through being the owner
		vmByOwner, err := vm_repo.New(version.V2).WithOwner(c.V2.Auth().User.ID).ExistsByID(id)
		if err != nil {
			return false, makeError(err)
		}

		if vmByOwner {
			return true, nil
		}

		// 3. User has access through a team
		teamAccess, err := c.V1.Teams().CheckResourceAccess(c.V2.Auth().User.ID, id)
		if err != nil {
			return false, makeError(err)
		}

		if teamAccess {
			return true, nil
		}

		return false, nil
	}

	// 4. No auth info was provided, always return true
	return true, nil
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
		log.Println("VM", id, "not found when repairing. Assuming it was deleted")
		return nil
	}

	err = c.K8s().Repair(id)
	if err != nil {
		return makeError(err)
	}

	log.Println("Repaired VM", id)
	return nil
}

// DoAction performs an action on the VM.
func (c *Client) DoAction(id string, dtoAction *body.VmActionCreate) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to perform action on vm %s. details: %w", id, err)
	}

	params := model.VmActionParams{}.FromDTOv2(dtoAction)

	vm, err := c.VM(id, nil)
	if err != nil {
		return makeError(err)
	}

	if vm == nil {
		log.Println("VM", id, "not found when performing action. Assuming it was deleted")
		return nil
	}

	err = c.K8s().DoAction(id, &params)
	if err != nil {
		return makeError(err)
	}

	log.Println("Performed action", params.Action, "on vm", id)
	return nil
}

// UpdateOwner updates the owner of the VM.
//
// This is the second step of the owner update process, where the transfer is actually done.
//
// It returns an error if the VM is not found.
func (c *Client) UpdateOwner(id string, params *model.VmUpdateOwnerParams) error {
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
		log.Println("VM", id, "not found when updating owner. Assuming it was deleted")
		return nil
	}

	err = vm_repo.New(version.V2).UpdateWithParams(id, &model.VmUpdateParams{
		OwnerID: &params.NewOwnerID,
	})
	if err != nil {
		return makeError(err)
	}

	_, err = c.Refresh(id)
	if err != nil {
		return makeError(err)
	}

	err = gpu_lease_repo.New().WithVmID(id).Release()
	if err != nil {
		return makeError(err)
	}

	err = c.K8s().EnsureOwner(id, params.OldOwnerID)
	if err != nil {
		return makeError(err)
	}

	// Restart VM to ensure possibly new specs are applied
	err = c.K8s().DoAction(id, &model.VmActionParams{Action: model.ActionRestart})
	if err != nil {
		return makeError(err)
	}

	nmc := notification_repo.New().FilterContent("id", id).WithType(model.NotificationVmTransfer)
	err = nmc.MarkReadAndCompleted()
	if err != nil {
		return makeError(err)
	}

	log.Println("VM", id, "owner updated from", params.OldOwnerID, " to", params.NewOwnerID)
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

	zone := config.Config.GetZone(vm.Zone)
	if zone == nil {
		return nil, makeError(sErrors.ZoneNotFoundErr)
	}

	var sshConnectionString *string
	if service := vm.Subsystems.K8s.GetService(fmt.Sprintf("%s-priv-22-prot-tcp", vm.Name)); service != nil {
		for _, port := range service.Ports {
			if port.TargetPort == 22 {
				sshConnectionString = utils.StrPtr(fmt.Sprintf("ssh root@%s -p %d", strings.Split(zone.Domains.ParentVM, ":")[0], port.Port))
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
func (c *Client) CheckQuota(id, userID string, quota *model.Quotas, opts ...opts.QuotaOpts) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check quota for user %s. details: %w", userID, err)
	}

	if !c.V2.HasAuth() || c.V2.Auth().User.IsAdmin {
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
		totalCpuCores := float64(usage.CpuCores + o.Create.CpuCores)
		totalRam := float64(usage.RAM + o.Create.RAM)
		totalDiskSize := float64(usage.DiskSize + o.Create.DiskSize)

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

		vm, err := vm_repo.New(version.V2).GetByID(id)
		if err != nil {
			return makeError(err)
		}

		if vm == nil {
			return makeError(sErrors.VmNotFoundErr)
		}

		if o.Update.CpuCores != nil {
			totalCpuCores := float64(usage.CpuCores)
			if o.Update.CpuCores != nil {
				totalCpuCores += float64(*o.Update.CpuCores - vm.Specs.CpuCores)
			}

			if totalCpuCores > quota.CpuCores {
				return sErrors.NewQuotaExceededError(fmt.Sprintf("CPU cores quota exceeded. Current: %d, Quota: %d", totalCpuCores, quota.CpuCores))
			}
		}

		if o.Update.RAM != nil {
			totalRam := float64(usage.RAM)
			if o.Update.RAM != nil {
				totalRam += float64(*o.Update.RAM - vm.Specs.RAM)
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
func (c *Client) GetUsage(userID string) (*model.VmUsage, error) {
	return vm_repo.New(version.V2).WithOwner(userID).GetUsage()
}

// GetHost gets the host for the VM.
//
// It returns an error if the VM is not found.
func (c *Client) GetHost(vmID string) (*model.Host, error) {
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
