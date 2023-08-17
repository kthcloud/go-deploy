package internal_service

import (
	"errors"
	"fmt"
	vmModel "go-deploy/models/sys/vm"
	gpuModel "go-deploy/models/sys/vm/gpu"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/cs"
	"go-deploy/pkg/subsystems/cs/commands"
	csModels "go-deploy/pkg/subsystems/cs/models"
	"log"
	"reflect"
	"strings"
	"time"
)

type CsCreated struct {
	VM *csModels.VmPublic
}

func withCsClient() (*cs.Client, error) {
	return cs.New(&cs.ClientConf{
		URL:         conf.Env.CS.URL,
		ApiKey:      conf.Env.CS.ApiKey,
		Secret:      conf.Env.CS.Secret,
		IpAddressID: conf.Env.CS.IpAddressID,
		NetworkID:   conf.Env.CS.NetworkID,
		ProjectID:   conf.Env.CS.ProjectID,
		ZoneID:      conf.Env.CS.ZoneID,
	})
}

func createServiceOfferingPublic(name string, cpuCores, ram, diskSize int) *csModels.ServiceOfferingPublic {
	return &csModels.ServiceOfferingPublic{
		Name:        name,
		Description: fmt.Sprintf("Auto-generated by deploy for vm %s", name),
		CpuCores:    cpuCores,
		RAM:         ram,
		DiskSize:    diskSize,
	}
}

func createCsVmPublic(name, serviceOfferingID, templateID string, tags []csModels.Tag) *csModels.VmPublic {
	return &csModels.VmPublic{
		Name:              name,
		ServiceOfferingID: serviceOfferingID,
		TemplateID:        templateID,
		ExtraConfig:       "",
		Tags:              tags,
	}
}

func createPortForwardingRulePublic(name, csVmID string, publicPort int, privatePort int, protocol string, tags []csModels.Tag) *csModels.PortForwardingRulePublic {
	return &csModels.PortForwardingRulePublic{
		Name:        name,
		VmID:        csVmID,
		PublicPort:  publicPort,
		PrivatePort: privatePort,
		Protocol:    protocol,
		Tags:        tags,
	}
}

func CreateCS(params *vmModel.CreateParams) (*CsCreated, error) {
	log.Println("setting up cs for", params.Name)

	userSshPublicKey := params.SshPublicKey
	adminSshPublicKey := conf.Env.VM.AdminSshPublicKey

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup cs for vm %s. details: %s", params.Name, err)
	}

	client, err := withCsClient()
	if err != nil {
		return nil, makeError(err)
	}

	vm, err := vmModel.GetByName(params.Name)
	if err != nil {
		return nil, makeError(err)
	}

	if vm == nil {
		log.Println("vm", params.Name, "not found for cs setup. assuming it was deleted")
		return nil, nil
	}

	// service offering
	serviceOffering := &vm.Subsystems.CS.ServiceOffering
	if !serviceOffering.Created() {
		serviceOffering, err = createServiceOffering(client, vm, createServiceOfferingPublic(params.Name, params.CpuCores, params.RAM, params.DiskSize))
		if err != nil {
			return nil, makeError(err)
		}
	}

	// vm
	csVM := &vm.Subsystems.CS.VM
	if !csVM.Created() {
		temporaryTemplateID := "cbac58b6-336b-49ab-b4d7-341586dfefcc"
		public := createCsVmPublic(params.Name, serviceOffering.ID, temporaryTemplateID, createDeployTags(params.Name, params.Name))

		csVM, err = createCsVM(client, vm, public, userSshPublicKey, adminSshPublicKey)
		if err != nil {
			// remove the service offering if the vm creation failed
			// however, do this as best-effort only to avoid cascading errors
			if serviceOffering.Created() {
				_ = client.DeleteServiceOffering(serviceOffering.ID)
				_ = vmModel.UpdateSubsystemByName(vm.Name, "cs", "serviceOffering", csModels.ServiceOfferingPublic{})
			}

			return nil, makeError(err)
		}
	}

	ruleMap := vm.Subsystems.CS.PortForwardingRuleMap
	if ruleMap == nil {
		ruleMap = map[string]csModels.PortForwardingRulePublic{}
	}

	for _, port := range params.Ports {
		var rule *csModels.PortForwardingRulePublic
		ruleInMap, hasRule := ruleMap[port.Name]
		if hasRule {
			rule = &ruleInMap
		}

		if !hasRule || !rule.Created() {
			freePort, err := client.GetFreePort(conf.Env.CS.PortRange.Start, conf.Env.CS.PortRange.End)
			if err != nil {
				return nil, makeError(err)
			}

			if freePort == -1 {
				return nil, makeError(fmt.Errorf("no free port found"))
			}

			public := createPortForwardingRulePublic(port.Name, csVM.ID, freePort, port.Port, port.Protocol, createDeployTags(port.Name, csVM.Name))
			rule, err = createPortForwardingRule(client, vm, port.Name, public)
			if err != nil {
				return nil, makeError(err)
			}
		}
	}

	return &CsCreated{
		VM: csVM,
	}, nil
}

func DeleteCS(name string) error {
	log.Println("deleting cs for", name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to delete cs for vm %s. details: %s", name, err)
	}

	client, err := withCsClient()
	if err != nil {
		return makeError(err)
	}

	vm, err := vmModel.GetByName(name)
	if err != nil {
		return makeError(err)
	}

	if vm == nil {
		log.Println("vm", name, "not found for cs deletion. assuming it was deleted")
		return nil
	}

	ruleMap := vm.Subsystems.CS.PortForwardingRuleMap

	for _, rule := range ruleMap {
		err = client.DeletePortForwardingRule(rule.ID)
		if err != nil {
			return makeError(err)
		}
	}

	err = vmModel.UpdateSubsystemByName(name, "cs", "portForwardingRuleMap", map[string]csModels.PortForwardingRulePublic{})
	if err != nil {
		return makeError(err)
	}

	if vm.Subsystems.CS.VM.ID != "" {
		err = client.DeleteVM(vm.Subsystems.CS.VM.ID)
		if err != nil {
			return makeError(err)
		}

		err = vmModel.UpdateSubsystemByName(name, "cs", "vm", csModels.VmPublic{})
		if err != nil {
			return makeError(err)
		}
	}

	if vm.Subsystems.CS.ServiceOffering.ID != "" {
		err = client.DeleteServiceOffering(vm.Subsystems.CS.ServiceOffering.ID)
		if err != nil {
			return makeError(err)
		}

		err = vmModel.UpdateSubsystemByName(name, "cs", "serviceOffering", csModels.ServiceOfferingPublic{})
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

func UpdateCS(vmID string, updateParams *vmModel.UpdateParams) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update cs for vm %s. details: %s", vmID, err)
	}

	client, err := withCsClient()
	if err != nil {
		return makeError(err)
	}

	vm, err := vmModel.GetByID(vmID)
	if err != nil {
		return makeError(err)
	}

	if vm == nil {
		log.Println("vm", vmID, "not found for cs update. assuming it was deleted")
		return nil
	}

	if vm.Subsystems.CS.VM.ID == "" {
		return nil
	}

	// port-forwarding rule
	if updateParams.Ports != nil {
		ruleMap := vm.Subsystems.CS.PortForwardingRuleMap
		if ruleMap == nil {
			ruleMap = make(map[string]csModels.PortForwardingRulePublic)
		}

		newRuleMap := make(map[string]csModels.PortForwardingRulePublic)
		createNewRuleMap := make(map[string]csModels.PortForwardingRulePublic)

		for _, port := range *updateParams.Ports {
			cmp1 := createPortForwardingRulePublic(port.Name, vm.Subsystems.CS.VM.ID, 0, port.Port, port.Protocol, createDeployTags(port.Name, vm.Name))
			cmp2, ok := ruleMap[port.Name]
			if !ok {
				// new rule
				createNewRuleMap[port.Name] = *cmp1
				delete(ruleMap, port.Name)
				continue
			}

			if cmp1.PrivatePort != cmp2.PrivatePort ||
				cmp1.Protocol != cmp2.Protocol ||
				cmp1.VmID != cmp2.VmID ||
				cmp2.ID == "" ||
				cmp2.PublicPort == 0 {
				err = client.DeletePortForwardingRule(cmp2.ID)
				if err != nil {
					return makeError(err)
				}

				createNewRuleMap[port.Name] = *cmp1
			} else {
				newRuleMap[port.Name] = cmp2
			}
			delete(ruleMap, port.Name)
		}

		// delete any remaining rules
		for _, rule := range ruleMap {
			err = client.DeletePortForwardingRule(rule.ID)
			if err != nil {
				return makeError(err)
			}
		}

		for name, public := range createNewRuleMap {
			var freePort int
			freePort, err = client.GetFreePort(conf.Env.CS.PortRange.Start, conf.Env.CS.PortRange.End)
			if err != nil {
				return makeError(err)
			}

			if freePort == -1 {
				return makeError(fmt.Errorf("no free port found"))
			}

			public.PublicPort = freePort

			rule, err := createPortForwardingRule(client, vm, name, &public)
			if err != nil {
				return makeError(err)
			}

			createNewRuleMap[name] = *rule
		}

		// merge the new and the old ones
		for name, public := range createNewRuleMap {
			newRuleMap[name] = public
		}

		err = vmModel.UpdateSubsystemByName(vm.Name, "cs", "portForwardingRuleMap", newRuleMap)
	}

	// service offering
	var serviceOffering *csModels.ServiceOfferingPublic
	if vm.Subsystems.CS.ServiceOffering.ID != "" {
		requiresUpdate := updateParams.CpuCores != nil && *updateParams.CpuCores != vm.Subsystems.CS.ServiceOffering.CpuCores ||
			updateParams.RAM != nil && *updateParams.RAM != vm.Subsystems.CS.ServiceOffering.RAM

		if requiresUpdate {
			err = client.DeleteServiceOffering(vm.Subsystems.CS.ServiceOffering.ID)
			if err != nil {
				return makeError(err)
			}

			id, err := client.CreateServiceOffering(&csModels.ServiceOfferingPublic{
				Name:        vm.Name,
				Description: vm.Subsystems.CS.ServiceOffering.Description,
				CpuCores:    *updateParams.CpuCores,
				RAM:         *updateParams.RAM,
				DiskSize:    vm.Specs.DiskSize,
			})
			if err != nil {
				return makeError(err)
			}

			serviceOffering, err = client.ReadServiceOffering(id)
			if err != nil {
				return makeError(err)
			}

			if serviceOffering == nil {
				return makeError(fmt.Errorf("failed to read service offering after creation"))
			}

			err = vmModel.UpdateSubsystemByName(vm.Name, "cs", "serviceOffering", *serviceOffering)
			if err != nil {
				return makeError(err)
			}
		} else {
			serviceOffering = &vm.Subsystems.CS.ServiceOffering
		}

		// make sure the vm is using the latest service offering
		if vm.Subsystems.CS.VM.ServiceOfferingID != serviceOffering.ID {
			vm.Subsystems.CS.VM.ServiceOfferingID = serviceOffering.ID

			// turn it off if it is on, but remember the status
			status, err := client.GetVmStatus(vm.Subsystems.CS.VM.ID)
			if err != nil {
				return makeError(err)
			}

			if status == "Running" {
				err = client.DoVmCommand(vm.Subsystems.CS.VM.ID, nil, "stop")
				if err != nil {
					return makeError(err)
				}
			}

			// update the service offering
			err = client.UpdateVM(&vm.Subsystems.CS.VM)
			if err != nil {
				return makeError(err)
			}

			err = vmModel.UpdateSubsystemByName(vm.Name, "cs", "vm", vm.Subsystems.CS.VM)
			if err != nil {
				return makeError(err)
			}

			// turn it on if it was on
			if status == "Running" {
				var requiredHost *string
				if vm.GpuID != "" {
					requiredHost, err = getRequiredHost(vm.GpuID)
					if err != nil {
						return makeError(err)
					}
				}

				err = client.DoVmCommand(vm.Subsystems.CS.VM.ID, requiredHost, "start")
				if err != nil {
					return makeError(err)
				}
			}

		}
	}

	return nil
}

func RepairCS(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to repair cs %s. details: %s", name, err)
	}

	vm, err := vmModel.GetByName(name)
	if err != nil {
		return makeError(err)
	}

	if vm == nil {
		log.Println("vm", name, "not found when repairing cs, assuming it was deleted")
		return nil
	}

	client, err := withCsClient()
	if err != nil {
		return makeError(err)
	}

	ss := vm.Subsystems.CS

	// service offering
	serviceOffering, err := client.ReadServiceOffering(ss.ServiceOffering.ID)
	if err != nil {
		return makeError(err)
	}

	if serviceOffering == nil || !reflect.DeepEqual(ss.ServiceOffering, *serviceOffering) {
		err = recreateServiceOffering(client, vm, &ss.ServiceOffering)
		if err != nil {
			return makeError(err)
		}
	}

	// vm
	csVM, err := client.ReadVM(ss.VM.ID)
	if err != nil {
		return makeError(err)
	}

	// this case must be handled manually, since we can't just recreate a virtual machine
	// thus, we don't care about every possible inconsistency, but only those that can be fixed by an update

	if csVM == nil {
		log.Println("vm", name, "not found when repairing cs, can't recreate a new one without a user ssh key")
		return nil
	}

	if csVM.ServiceOfferingID != ss.ServiceOffering.ID ||
		!reflect.DeepEqual(csVM.Tags, ss.VM.Tags) ||
		csVM.Name != ss.VM.Name {
		err = client.UpdateVM(&ss.VM)
		if err != nil {
			return makeError(err)
		}
	}

	ruleMap := &ss.PortForwardingRuleMap
	if ruleMap != nil {
		for name, dbRule := range *ruleMap {
			var rule *csModels.PortForwardingRulePublic
			rule, err = client.ReadPortForwardingRule(dbRule.ID)
			if err != nil {
				return makeError(err)
			}

			if rule == nil || !reflect.DeepEqual(dbRule, *rule) {
				err = recreatePortForwardingRule(client, vm, name, &dbRule)
				if err != nil {
					return makeError(err)
				}
			}
		}
	}

	return nil
}

func AttachGPU(gpuID, vmID string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to attach gpu %s to cs vm %s. details: %s", gpuID, vmID, err)
	}

	client, err := withCsClient()
	if err != nil {
		return makeError(err)
	}

	vm, err := vmModel.GetByID(vmID)
	if err != nil {
		return makeError(err)
	}

	if vm == nil {
		log.Println("vm", vmID, "not found when attaching gpu", gpuID, "to cs vm, assuming it was deleted")
		return nil
	}

	if vm.Subsystems.CS.VM.ID == "" {
		log.Println("vm", vmID, "has no cs vm id when attaching gpu", gpuID, "to cs vm, assuming it was deleted")
		return nil
	}

	gpu, err := gpuModel.GetByID(gpuID)
	if err != nil {
		return makeError(err)
	}

	requiredExtraConfig := createExtraConfig(gpu)
	currentExtraConfig := vm.Subsystems.CS.VM.ExtraConfig
	if requiredExtraConfig != currentExtraConfig {
		var status string
		status, err = client.GetVmStatus(vm.Subsystems.CS.VM.ID)
		if err != nil {
			return makeError(err)
		}

		if status == "Running" {
			err = client.DoVmCommand(vm.Subsystems.CS.VM.ID, nil, "stop")
			if err != nil {
				return makeError(err)
			}
		}

		vm.Subsystems.CS.VM.ExtraConfig = requiredExtraConfig

		err = client.UpdateVM(&vm.Subsystems.CS.VM)
		if err != nil {
			return makeError(err)
		}

		err = vmModel.UpdateSubsystemByName(vm.Name, "cs", "vm.extraConfig", vm.Subsystems.CS.VM.ExtraConfig)
		if err != nil {
			return makeError(err)
		}
	}

	// always start the vm after attaching gpu, to make sure the vm can be started on the host
	requiredHost, err := getRequiredHost(gpuID)
	if err != nil {
		return makeError(err)
	}

	err = client.DoVmCommand(vm.Subsystems.CS.VM.ID, requiredHost, "start")
	if err != nil {
		return makeError(err)
	}

	return nil
}

func DetachGPU(vmID string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to detach gpu from cs vm %s. details: %s", vmID, err)
	}

	client, err := withCsClient()
	if err != nil {
		return makeError(err)
	}

	vm, err := vmModel.GetByID(vmID)
	if err != nil {
		return makeError(err)
	}

	if vm == nil {
		log.Println("vm", vmID, "not found for when detaching gpu in cs. assuming it was deleted")
		return nil
	}

	if vm.Subsystems.CS.VM.ID == "" {
		return nil
	}

	// turn it off if it is on, but remember the status
	status, err := client.GetVmStatus(vm.Subsystems.CS.VM.ID)
	if err != nil {
		return makeError(err)
	}

	if status == "Running" {
		err = client.DoVmCommand(vm.Subsystems.CS.VM.ID, nil, "stop")
		if err != nil {
			return makeError(err)
		}
	}

	vm.Subsystems.CS.VM.ExtraConfig = ""

	err = client.UpdateVM(&vm.Subsystems.CS.VM)
	if err != nil {
		return makeError(err)
	}

	err = vmModel.UpdateSubsystemByName(vm.Name, "cs", "vm.extraConfig", vm.Subsystems.CS.VM.ExtraConfig)
	if err != nil {
		return makeError(err)
	}

	// turn it on if it was on
	if status == "Running" {
		err = client.DoVmCommand(vm.Subsystems.CS.VM.ID, nil, "start")
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

func IsGpuAttachedCS(host string, bus string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if gpu %s:%s is attached to any cs vm. details: %s", host, bus, err)
	}

	client, err := withCsClient()
	if err != nil {
		return false, makeError(err)
	}

	params := client.CsClient.VirtualMachine.NewListVirtualMachinesParams()
	params.SetListall(true)

	vms, err := client.CsClient.VirtualMachine.ListVirtualMachines(params)
	if err != nil {
		return false, makeError(err)
	}

	for _, vm := range vms.VirtualMachines {
		if vm.Details != nil && vm.Hostname == host {
			extraConfig, ok := vm.Details["extraconfig-1"]
			if ok {
				if strings.Contains(extraConfig, fmt.Sprintf("bus='0x%s'", bus)) {
					return true, nil
				}
			}
		}
	}

	return false, nil
}

func CreateSnapshotCS(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create snapshot for cs vm %s. details: %s", id, err)
	}

	client, err := withCsClient()
	if err != nil {
		return makeError(err)
	}

	vm, err := vmModel.GetByID(id)
	if err != nil {
		return makeError(err)
	}

	if vm == nil {
		log.Println("vm", id, "not found for when creating snapshot in cs. assuming it was deleted")
		return nil
	}

	snapshotMap := vm.Subsystems.CS.SnapshotMap
	if snapshotMap == nil {
		snapshotMap = map[string]csModels.SnapshotPublic{}
	}

	name := fmt.Sprintf("snapshot-%s", time.Now().Format("20060102150405"))
	if _, ok := snapshotMap[name]; ok {
		log.Println("snapshot", name, "already exists for vm", id)
		return nil
	}

	vmStatus, err := client.GetVmStatus(vm.Subsystems.CS.VM.ID)
	if err != nil {
		return makeError(err)
	}

	if vmStatus != "Running" {
		return fmt.Errorf("vm %s is not running", id)
	}

	public := &csModels.SnapshotPublic{
		Name: name,
		VmID: vm.Subsystems.CS.VM.ID,
	}

	snapshotID, err := client.CreateSnapshot(public)
	if err != nil {
		return makeError(err)
	}

	log.Println("created snapshot", snapshotID, "for vm", id)

	return nil
}

func DoCommandCS(vmID string, gpuID *string, command string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to execute command %s for cs vm %s. details: %s", command, vmID, err)
	}

	client, err := withCsClient()
	if err != nil {
		return makeError(err)
	}

	var requiredHost *string
	if gpuID != nil {
		requiredHost, err = getRequiredHost(*gpuID)
		if err != nil {
			return makeError(err)
		}
	}

	err = client.DoVmCommand(vmID, requiredHost, commands.Command(command))
	if err != nil {
		return makeError(err)
	}

	return nil
}

func CanStartCS(vmID, hostName string) (bool, string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if cs vm %s can be started on host %s. details: %s", vmID, hostName, err)
	}

	client, err := withCsClient()
	if err != nil {
		return false, "", makeError(err)
	}

	hasCapacity, err := client.HasCapacity(vmID, hostName)
	if err != nil {
		return false, "", err
	}

	if !hasCapacity {
		return false, "Host doesn't have capacity", nil
	}

	correctState, reason, err := HostInCorrectState(hostName)
	if err != nil {
		return false, "", err
	}

	if !correctState {
		return false, reason, nil
	}

	return true, "", nil
}

func HostInCorrectState(hostName string) (bool, string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if host %s is in correct state. details: %s", hostName, err)
	}

	client, err := withCsClient()
	if err != nil {
		return false, "", makeError(err)
	}

	host, err := client.ReadHostByName(hostName)
	if err != nil {
		return false, "", makeError(err)
	}

	if host.State != "Up" {
		return false, "Host is not up", nil
	}

	if host.ResourceState != "Enabled" {
		return false, "Host is not enabled", nil
	}

	return true, "", nil
}

func recreateServiceOffering(client *cs.Client, vm *vmModel.VM, public *csModels.ServiceOfferingPublic) error {
	err := client.DeleteServiceOffering(vm.Subsystems.CS.ServiceOffering.ID)
	if err != nil {
		return err
	}

	_, err = createServiceOffering(client, vm, public)
	if err != nil {
		return err
	}

	return nil
}

func recreatePortForwardingRule(client *cs.Client, vm *vmModel.VM, name string, public *csModels.PortForwardingRulePublic) error {
	rule, ok := vm.Subsystems.CS.PortForwardingRuleMap[name]
	if !ok {
		err := client.DeletePortForwardingRule(rule.ID)
		if err != nil {
			return err
		}
	}

	_, err := createPortForwardingRule(client, vm, name, public)
	if err != nil {
		return err
	}

	return nil
}

func createServiceOffering(client *cs.Client, vm *vmModel.VM, public *csModels.ServiceOfferingPublic) (*csModels.ServiceOfferingPublic, error) {
	id, err := client.CreateServiceOffering(public)
	if err != nil {
		return nil, err
	}

	serviceOffering, err := client.ReadServiceOffering(id)
	if err != nil {
		return nil, err
	}

	if serviceOffering == nil {
		return nil, errors.New("failed to read service offering after creation")
	}

	err = vmModel.UpdateSubsystemByName(vm.Name, "cs", "serviceOffering", serviceOffering)
	if err != nil {
		return nil, err
	}

	return serviceOffering, nil
}

func createCsVM(client *cs.Client, vm *vmModel.VM, public *csModels.VmPublic, userSshKey, adminSshKey string) (*csModels.VmPublic, error) {
	id, err := client.CreateVM(public, userSshKey, adminSshKey)
	if err != nil {
		return nil, err
	}

	csVM, err := client.ReadVM(id)
	if err != nil {
		return nil, err
	}

	if csVM == nil {
		return nil, errors.New("failed to read vm after creation")
	}

	err = vmModel.UpdateSubsystemByName(vm.Name, "cs", "vm", csVM)
	if err != nil {
		return nil, err
	}

	vm.Subsystems.CS.VM = *csVM

	return csVM, nil
}

func createPortForwardingRule(client *cs.Client, vm *vmModel.VM, name string, public *csModels.PortForwardingRulePublic) (*csModels.PortForwardingRulePublic, error) {
	id, err := client.CreatePortForwardingRule(public)
	if err != nil {
		return nil, err
	}

	portForwardingRule, err := client.ReadPortForwardingRule(id)
	if err != nil {
		return nil, err
	}

	if portForwardingRule == nil {
		return nil, errors.New("failed to read port forwarding rule after creation")
	}

	if vm.Subsystems.CS.PortForwardingRuleMap == nil {
		vm.Subsystems.CS.PortForwardingRuleMap = make(map[string]csModels.PortForwardingRulePublic)
	}
	vm.Subsystems.CS.PortForwardingRuleMap[name] = *portForwardingRule

	err = vmModel.UpdateSubsystemByName(vm.Name, "cs", "portForwardingRuleMap", vm.Subsystems.CS.PortForwardingRuleMap)
	if err != nil {
		return nil, err
	}

	return portForwardingRule, nil
}

func createDeployTags(name string, deployName string) []csModels.Tag {
	return []csModels.Tag{
		{Key: "name", Value: name},
		{Key: "managedBy", Value: conf.Env.Manager},
		{Key: "deployName", Value: deployName},
	}
}

func createExtraConfig(gpu *gpuModel.GPU) string {
	data := fmt.Sprintf(`
<devices> <hostdev mode='subsystem' type='pci' managed='yes'> <driver name='vfio' />
	<source> <address domain='0x0000' bus='0x%s' slot='0x00' function='0x0' /> </source> 
	<alias name='nvidia0' /> <address type='pci' domain='0x0000' bus='0x00' slot='0x00' function='0x0' /> 
</hostdev> </devices>`, gpu.Data.Bus)

	data = strings.Replace(data, "\n", "", -1)
	data = strings.Replace(data, "\t", "", -1)

	return data
}

func getRequiredHost(gpuID string) (*string, error) {
	gpu, err := gpuModel.GetByID(gpuID)
	if err != nil {
		return nil, err
	}

	if gpu.Host == "" {
		return nil, fmt.Errorf("no host found for gpu %s", gpu.ID)
	}

	return &gpu.Host, nil
}
