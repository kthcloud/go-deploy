package vm_service

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go-deploy/models/dto/body"
	"go-deploy/models/sys/gpu"
	jobModel "go-deploy/models/sys/job"
	notificationModel "go-deploy/models/sys/notification"
	roleModel "go-deploy/models/sys/role"
	teamModels "go-deploy/models/sys/team"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/pkg/config"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/job_service"
	"go-deploy/service/notification_service"
	"go-deploy/service/vm_service/client"
	"go-deploy/service/vm_service/cs_service"
	"go-deploy/service/vm_service/k8s_service"
	"go-deploy/utils"
	"go.mongodb.org/mongo-driver/bson"
	"log"
	"sort"
	"strings"
)

// Get gets an existing deployment.
//
// It can be fetched in multiple ways including ID, name, transfer code, and Harbor webhook.
// It supports service.AuthInfo, and will restrict the result to ensure the user has access to the resource.
func (c *Client) Get(id string, opts *client.GetOptions) (*vmModel.VM, error) {
	vClient := vmModel.New()

	if opts.TransferCode != "" {
		return vClient.GetByTransferCode(opts.TransferCode)
	}

	var effectiveUserID string
	if c.Auth != nil && c.Auth.IsAdmin {
		effectiveUserID = c.Auth.UserID
	}

	var teamCheck bool
	if !opts.Shared {
		teamCheck = false
	} else if c.Auth == nil || c.Auth.IsAdmin {
		teamCheck = true
	} else {
		var err error
		teamCheck, err = teamModels.New().AddUserID(c.Auth.UserID).AddResourceID(id).ExistsAny()
		if err != nil {
			return nil, err
		}
	}

	if !teamCheck && effectiveUserID != "" {
		vClient.RestrictToOwner(effectiveUserID)
	}

	return c.VM(id, vClient)
}

// List lists existing deployments.
//
// It supports service.AuthInfo, and will restrict the result to ensure the user has access to the resource.
func (c *Client) List(opts *client.ListOptions) ([]vmModel.VM, error) {
	vClient := vmModel.New()

	if opts.Pagination != nil {
		vClient.WithPagination(opts.Pagination.Page, opts.Pagination.PageSize)
	}

	var effectiveUserID string
	if opts.UserID != "" {
		// specific user's deployments are requested
		if c.Auth == nil || c.Auth.UserID == opts.UserID || c.Auth.IsAdmin {
			effectiveUserID = opts.UserID
		} else {
			effectiveUserID = c.Auth.UserID
		}
	} else {
		// all deployments are requested
		if c.Auth != nil && !c.Auth.IsAdmin {
			effectiveUserID = c.Auth.UserID
		}
	}

	if effectiveUserID != "" {
		vClient.RestrictToOwner(effectiveUserID)
	}

	resources, err := c.VMs(vClient)
	if err != nil {
		return nil, err
	}

	if opts.Shared && effectiveUserID != "" {
		skipIDs := make([]string, len(resources))
		for i, resource := range resources {
			skipIDs[i] = resource.ID
		}

		teamClient := teamModels.New().AddUserID(effectiveUserID)
		if opts.Pagination != nil {
			teamClient.WithPagination(opts.Pagination.Page, opts.Pagination.PageSize)
		}

		teams, err := teamClient.List()
		if err != nil {
			return nil, err
		}

		for _, team := range teams {
			for _, resource := range team.GetResourceMap() {
				if resource.Type != teamModels.ResourceTypeDeployment {
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
				if err != nil && vm != nil {
					resources = append(resources, *vm)
				}
			}
		}

		sort.Slice(resources, func(i, j int) bool {
			return resources[i].CreatedAt.After(resources[j].CreatedAt)
		})

		// since we fetched from two collections, we need to do pagination manually
		if opts.Pagination != nil {
			resources = utils.GetPage(resources, opts.Pagination.PageSize, opts.Pagination.Page)
		}

	} else {
		// sort by createdAt
		sort.Slice(resources, func(i, j int) bool {
			return resources[i].CreatedAt.After(resources[j].CreatedAt)
		})
	}

	return resources, nil
}

func (c *Client) Create(id, ownerID string, dtoVmCreate *body.VmCreate) error {
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
			return sErrors.NonUniqueFieldErr
		}

		return makeError(err)
	}

	err = cs_service.New(c.Context).Create(id, params)
	if err != nil {
		return makeError(err)
	}

	// there is always at least one port: __ssh
	if len(params.Ports) > 1 {
		err = k8s_service.New(c.Context).Create(id, params)
		if err != nil {
			return makeError(err)
		}
	} else {
		log.Println("skipping k8s setup for vm", id, "since it has no ports")
	}

	return nil
}

func (c *Client) Update(id string, dtoVmUpdate *body.VmUpdate) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update vm. details: %w", err)
	}

	vmUpdate := &vmModel.UpdateParams{}
	vmUpdate.FromDTO(dtoVmUpdate)

	if vmUpdate.SnapshotID != nil {
		err := c.ApplySnapshot(id, *vmUpdate.SnapshotID)
		if err != nil {
			return makeError(err)
		}
	} else {
		err := vmModel.New().UpdateWithParamsByID(id, vmUpdate)
		if err != nil {
			if errors.Is(err, vmModel.NonUniqueFieldErr) {
				return sErrors.NonUniqueFieldErr
			}

			return makeError(err)
		}

		err = cs_service.New(c.Context).Update(id, vmUpdate)
		if err != nil {
			return makeError(err)
		}

		err = k8s_service.New(c.Context).Repair(id)
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

func (c *Client) Delete(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete vm. details: %w", err)
	}

	vm, err := c.VM(id, nil)
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

	err = k8s_service.New(c.Context).Delete(id)
	if err != nil {
		return makeError(err)
	}

	err = cs_service.New(c.Context).Delete(vm.ID)
	if err != nil {
		return makeError(err)
	}

	err = gpu.New().Detach(vm.ID)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func (c *Client) Repair(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to repair vm %s. details: %w", id, err)
	}

	vm, err := c.Get(id, &client.GetOptions{})
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

	err = cs_service.New(c.Context).Repair(id)
	if err != nil {
		return makeError(err)
	}

	err = k8s_service.New(c.Context).Repair(id)
	if err != nil {
		return makeError(err)
	}

	log.Println("repaired vm", id)
	return nil
}

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

	if c.Auth == nil || c.Auth.IsAdmin {
		transferDirectly = true
	} else if c.Auth.UserID == params.NewOwnerID {
		if params.TransferCode == nil || vm.Transfer == nil || vm.Transfer.Code != *params.TransferCode {
			return nil, sErrors.InvalidTransferCodeErr
		}

		transferDirectly = true
	}

	if transferDirectly {
		jobID := uuid.New().String()
		err := job_service.Create(jobID, c.Auth.UserID, jobModel.TypeUpdateVmOwner, map[string]interface{}{
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
			"email":  c.Auth.GetEmail(),
			"code":   code,
		},
	})

	if err != nil {
		return nil, makeError(err)
	}

	return nil, nil
}

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

	err = vmModel.New().UpdateWithParamsByID(id, &vmModel.UpdateParams{
		OwnerID:        &params.NewOwnerID,
		TransferCode:   &emptyString,
		TransferUserID: &emptyString,
	})
	if err != nil {
		return makeError(err)
	}

	err = gpu.New().WithVM(id).UpdateWithBson(bson.D{{"lease.user", params.NewOwnerID}})
	if err != nil {
		return makeError(err)
	}

	err = cs_service.New(c.Context).EnsureOwner(id, params.OldOwnerID)
	if err != nil {
		return makeError(err)
	}

	err = k8s_service.New(c.Context).EnsureOwner(id, params.OldOwnerID)
	if err != nil {
		return makeError(err)
	}

	log.Println("vm", id, "owner updated to", params.NewOwnerID)
	return nil
}

func (c *Client) ClearUpdateOwner(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to clear vm owner update. details: %w", err)
	}

	deployment, err := vmModel.New().GetByID(id)
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
	port := vm.Subsystems.CS.PortForwardingRuleMap["__ssh"].PublicPort

	if domainName == "" || port == 0 {
		return nil, nil
	}

	connectionString := fmt.Sprintf("ssh root@%s -p %d", domainName, port)

	return &connectionString, nil
}

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

		err = cs_service.New(c.Context).DoCommand(vm.ID, csID, vm.GetGpuID(), command)
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

	err = vmModel.New().AddActivity(id, activity)
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

func NameAvailable(name string) (bool, error) {
	exists, err := vmModel.New().ExistsByName(name)
	if err != nil {
		return false, err
	}

	return !exists, nil
}

func HttpProxyNameAvailable(id, name string) (bool, error) {
	filter := bson.D{
		{"id", bson.D{{"$ne", id}}},
		{"ports.httpProxy.name", name},
	}

	exists, err := vmModel.New().WithCustomFilter(filter).ExistsAny()
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
func (c *Client) CheckQuota(id, userID string, quota *roleModel.Quotas, opts *client.QuotaOptions) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check quota. details: %w", err)
	}

	if c.Auth == nil || c.Auth.IsAdmin {
		return nil
	}

	usage, err := c.GetUsage(userID)
	if err != nil {
		return makeError(err)
	}

	if usage == nil {
		return makeError(fmt.Errorf("failed to get usage for user %s", userID))
	}

	if opts.Create != nil {
		totalCpuCores := usage.CpuCores + opts.Create.CpuCores
		totalRam := usage.RAM + opts.Create.RAM
		totalDiskSize := usage.DiskSize + opts.Create.DiskSize

		if totalCpuCores > quota.CpuCores {
			return sErrors.NewQuotaExceededError(fmt.Sprintf("CPU cores quota exceeded. Current: %d, Quota: %d", totalCpuCores, quota.CpuCores))
		}

		if totalRam > quota.RAM {
			return sErrors.NewQuotaExceededError(fmt.Sprintf("RAM quota exceeded. Current: %d, Quota: %d", totalRam, quota.RAM))
		}

		if totalDiskSize > quota.DiskSize {
			return sErrors.NewQuotaExceededError(fmt.Sprintf("Disk size quota exceeded. Current: %d, Quota: %d", totalDiskSize, quota.DiskSize))
		}
	} else if opts.Update != nil {
		if opts.Update.CpuCores == nil && opts.Update.RAM == nil {
			return nil
		}

		vm, err := vmModel.New().GetByID(id)
		if err != nil {
			return makeError(err)
		}

		if vm == nil {
			return makeError(sErrors.VmNotFoundErr)
		}

		if opts.Update.CpuCores != nil {
			totalCpuCores := usage.CpuCores
			if opts.Update.CpuCores != nil {
				totalCpuCores += *opts.Update.CpuCores - vm.Specs.CpuCores
			}

			if totalCpuCores > quota.CpuCores {
				return sErrors.NewQuotaExceededError(fmt.Sprintf("CPU cores quota exceeded. Current: %d, Quota: %d", totalCpuCores, quota.CpuCores))
			}
		}

		if opts.Update.RAM != nil {
			totalRam := usage.RAM
			if opts.Update.RAM != nil {
				totalRam += *opts.Update.RAM - vm.Specs.RAM
			}

			if totalRam > quota.RAM {
				return sErrors.NewQuotaExceededError(fmt.Sprintf("RAM quota exceeded. Current: %d, Quota: %d", totalRam, quota.RAM))
			}
		}
	} else if opts.CreateSnapshot != nil {
		if usage.Snapshots >= quota.Snapshots {
			return sErrors.NewQuotaExceededError(fmt.Sprintf("Snapshot count quota exceeded. Current: %d, Quota: %d", usage.Snapshots, quota.Snapshots))
		}
	}

	return nil
}

func (c *Client) GetUsage(userID string) (*vmModel.Usage, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check quota. details: %w", err)
	}

	usage := &vmModel.Usage{}

	currentVms, err := vmModel.New().RestrictToOwner(userID).List()
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

func (c *Client) GetHost(vmID string) (*vmModel.Host, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get host for vm %s. details: %w", vmID, err)
	}

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

	cc := cs_service.New(c.Context)

	host, err := cc.GetHostByVM(vmID)
	if err != nil {
		return nil, makeError(err)
	}

	if host != nil {
		return &vmModel.Host{
			ID:   host.ID,
			Name: host.Name,
		}, nil
	}

	idStruct, err := gpu.New().WithVM(vmID).GetID()
	if err != nil {
		return nil, makeError(err)
	}

	if idStruct != nil {
		hostName, err := cc.GetRequiredHost(idStruct.ID)
		if err != nil {
			return nil, makeError(err)
		}

		if hostName != nil {
			host, err = cc.GetHostByName(*hostName, zone)
			if err != nil {
				return nil, makeError(err)
			}

			return &vmModel.Host{
				ID:   host.ID,
				Name: host.Name,
			}, nil
		}
	}

	return nil, nil
}

type CloudStackHostCapabilities struct {
	CpuCoresTotal int
	CpuCoresUsed  int
	RamTotal      int
	RamUsed       int
}

func GetCloudStackHostCapabilities(hostName string, zoneName string) (*CloudStackHostCapabilities, error) {
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

	return &CloudStackHostCapabilities{
		CpuCoresTotal: host.CpuCoresTotal * configuration.OverProvisioningFactor,
		CpuCoresUsed:  host.CpuCoresUsed,
		RamTotal:      host.RamTotal,
		RamUsed:       host.RamUsed,
	}, nil
}

func createTransferCode() string {
	return utils.HashString(uuid.NewString())
}
