package cs_service

import (
	"fmt"
	"go-deploy/models/sys/deployment/storage_manager"
	"go-deploy/models/sys/enviroment"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/cs/commands"
	csModels "go-deploy/pkg/subsystems/cs/models"
	"go-deploy/service"
	"go-deploy/service/vm_service/cs_service/helpers"
	"log"
	"reflect"
	"time"
)

// TemplateID this is a temporary solution until we have template selection
const TemplateID = "cbac58b6-336b-49ab-b4d7-341586dfefcc"

func CreateCS(params *vmModel.CreateParams) (*CsCreated, error) {
	log.Println("setting up cs for", params.Name)

	userSshPublicKey := params.SshPublicKey
	adminSshPublicKey := conf.Env.VM.AdminSshPublicKey

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup cs for vm %s. details: %w", params.Name, err)
	}

	zone := conf.Env.VM.GetZone(params.Zone)
	if zone == nil {
		return nil, makeError(fmt.Errorf("zone %s not found", params.Zone))
	}

	client, err := helpers.WithCsClient(zone)
	if err != nil {
		return nil, makeError(err)
	}

	vm, err := vmModel.New().GetByName(params.Name)
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
		public := helpers.CreateServiceOfferingPublic(params.Name, params.CpuCores, params.RAM, params.DiskSize)
		serviceOffering, err = helpers.CreateServiceOffering(client, vm, public, vmModel.New().UpdateSubsystemByName)
		if err != nil {
			return nil, makeError(err)
		}
	}

	// vm
	csVM := &vm.Subsystems.CS.VM
	if !csVM.Created() {
		public := helpers.CreateCsVmPublic(params.Name, serviceOffering.ID, TemplateID, helpers.CreateDeployTags(params.Name, params.Name))

		csVM, err = helpers.CreateCsVM(client, vm, public, userSshPublicKey, adminSshPublicKey, vmModel.New().UpdateSubsystemByName)
		if err != nil {
			// remove the service offering if the vm creation failed.
			// however, do this as best-effort only to avoid cascading errors
			if serviceOffering.Created() {
				_ = client.DeleteServiceOffering(serviceOffering.ID)
				_ = vmModel.New().UpdateSubsystemByName(vm.Name, "cs", "serviceOffering", csModels.ServiceOfferingPublic{})
			}

			return nil, makeError(err)
		}
	}

	ruleMap := vm.Subsystems.CS.GetPortForwardingRuleMap()
	if ruleMap == nil {
		ruleMap = map[string]csModels.PortForwardingRulePublic{}
	}

	for _, port := range params.Ports {
		if rule, hasRule := ruleMap[port.Name]; hasRule && rule.Created() {
			continue
		}

		freePort, err := client.GetFreePort(zone.PortRange.Start, zone.PortRange.End)
		if err != nil {
			return nil, makeError(err)
		}

		useRootNetwork := params.NetworkID == nil
		if useRootNetwork {
			rootPublic := helpers.CreateRootPortForwardingRulePublic(port.Name, csVM.ID, freePort, port.Port, port.Protocol, helpers.CreateDeployTags(port.Name, csVM.Name), *zone)
			_, err = helpers.CreatePortForwardingRule(client, vm, port.Name, rootPublic, vmModel.New().UpdateSubsystemByName)
			if err != nil {
				return nil, makeError(err)
			}
		} else {
			network, err := client.ReadNetwork(*params.NetworkID)
			if err != nil {
				return nil, makeError(err)
			}

			rootPublic := helpers.CreateRootPortForwardingRulePublic(
				port.Name,
				csVM.ID,
				freePort,
				port.Port,
				port.Protocol,
				helpers.CreateDeployTags(port.Name, csVM.Name), *zone,
			)
			_, err = helpers.CreatePortForwardingRule(client, vm, port.Name, rootPublic, vmModel.New().UpdateSubsystemByName)
			if err != nil {
				return nil, makeError(err)
			}

			ipAddressID, err := client.GetNetworkSourceNatIpAddressID(*params.NetworkID)
			if err != nil {
				return nil, makeError(err)
			}

			public := helpers.CreatePortForwardingRulePublic(
				helpers.CreateNonRootPortForwardingRuleName(port.Name, network.Name),
				csVM.ID,
				*params.NetworkID,
				ipAddressID,
				port.Port,
				port.Protocol,
				helpers.CreateDeployTags(port.Name, csVM.Name),
			)
			_, err = helpers.CreatePortForwardingRule(client, vm, port.Name, public, vmModel.New().UpdateSubsystemByName)
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
		return fmt.Errorf("failed to delete cs for vm %s. details: %w", name, err)
	}

	vm, err := vmModel.New().GetByName(name)
	if err != nil {
		return makeError(err)
	}

	if vm == nil {
		log.Println("vm", name, "not found for cs deletion. assuming it was deleted")
		return nil
	}

	zone := conf.Env.VM.GetZone(vm.Zone)
	if zone == nil {
		return makeError(fmt.Errorf("zone %s not found", vm.Zone))
	}

	client, err := helpers.WithCsClient(zone)
	if err != nil {
		return makeError(err)
	}

	ruleMap := vm.Subsystems.CS.GetPortForwardingRuleMap()

	for _, rule := range ruleMap {
		err = client.DeletePortForwardingRule(rule.ID)
		if err != nil {
			return makeError(err)
		}
	}

	err = vmModel.New().UpdateSubsystemByName(name, "cs", "portForwardingRuleMap", map[string]csModels.PortForwardingRulePublic{})
	if err != nil {
		return makeError(err)
	}

	if vm.Subsystems.CS.VM.ID != "" {
		err = client.DeleteVM(vm.Subsystems.CS.VM.ID)
		if err != nil {
			return makeError(err)
		}

		err = vmModel.New().UpdateSubsystemByName(name, "cs", "vm", csModels.VmPublic{})
		if err != nil {
			return makeError(err)
		}
	}

	if vm.Subsystems.CS.ServiceOffering.ID != "" {
		err = client.DeleteServiceOffering(vm.Subsystems.CS.ServiceOffering.ID)
		if err != nil {
			return makeError(err)
		}

		err = vmModel.New().UpdateSubsystemByName(name, "cs", "serviceOffering", csModels.ServiceOfferingPublic{})
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

func UpdateCS(vmID string, updateParams *vmModel.UpdateParams) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update cs for vm %s. details: %w", vmID, err)
	}

	vm, err := vmModel.New().GetByID(vmID)
	if err != nil {
		return makeError(err)
	}

	if vm == nil {
		log.Println("vm", vmID, "not found for cs update. assuming it was deleted")
		return nil
	}

	zone := conf.Env.VM.GetZone(vm.Zone)
	if zone == nil {
		return makeError(fmt.Errorf("zone %s not found", vm.Zone))
	}

	client, err := helpers.WithCsClient(zone)
	if err != nil {
		return makeError(err)
	}

	if !vm.Subsystems.CS.VM.Created() {
		log.Println("cs vm", vmID, "not created when updating cs. assuming it was deleted or not yet created")
		return nil
	}

	// port-forwarding rule
	if updateParams.Ports != nil {
		ruleMap := vm.Subsystems.CS.GetPortForwardingRuleMap()

		newRuleMap := make(map[string]csModels.PortForwardingRulePublic)
		createNewRuleMap := make(map[string]csModels.PortForwardingRulePublic)

		for _, port := range *updateParams.Ports {
			useRootNetwork := vm.NetworkID == nil
			cmp1 := helpers.CreateRootPortForwardingRulePublic(
				port.Name,
				vm.Subsystems.CS.VM.ID,
				// set this to zero here, in case we don't need a new rule (getting a free port is expensive)
				0, port.Port,
				port.Protocol,
				helpers.CreateDeployTags(port.Name, vm.Name),
				*zone,
			)
			cmp2, ok := ruleMap[port.Name]
			if !ok {
				// new rule
				createNewRuleMap[port.Name] = *cmp1
				delete(ruleMap, port.Name)
				continue
			}

			cmp2Cleaned := cmp2
			cmp2Cleaned.ID = ""
			cmp2Cleaned.CreatedAt = time.Time{}

			if !reflect.DeepEqual(cmp1, cmp2Cleaned) {
				err = client.DeletePortForwardingRule(cmp2.ID)
				if err != nil {
					return makeError(err)
				}

				createNewRuleMap[port.Name] = *cmp1
			} else {
				createNewRuleMap[port.Name] = cmp2
			}

			delete(ruleMap, port.Name)

			if !useRootNetwork {
				network, err := client.ReadNetwork(*vm.NetworkID)
				if err != nil {
					return makeError(err)
				}

				cmp1 = helpers.CreatePortForwardingRulePublic(
					helpers.CreateNonRootPortForwardingRuleName(port.Name, network.Name),
					vm.Subsystems.CS.VM.ID,
					*vm.NetworkID,
					cmp2.IpAddressID,
					port.Port,
					port.Protocol,
					helpers.CreateDeployTags(port.Name, vm.Name),
				)
				cmp2, ok = ruleMap[helpers.CreateNonRootPortForwardingRuleName(port.Name, network.Name)]
				if !ok {
					// new rule
					createNewRuleMap[helpers.CreateNonRootPortForwardingRuleName(port.Name, network.Name)] = *cmp1
					delete(ruleMap, helpers.CreateNonRootPortForwardingRuleName(port.Name, network.Name))
					continue
				}

				cmp2Cleaned = cmp2
				cmp2Cleaned.ID = ""
				cmp2Cleaned.CreatedAt = time.Time{}

				if !reflect.DeepEqual(cmp1, cmp2Cleaned) {
					err = client.DeletePortForwardingRule(cmp2.ID)
					if err != nil {
						return makeError(err)
					}

					createNewRuleMap[helpers.CreateNonRootPortForwardingRuleName(port.Name, network.Name)] = *cmp1
				} else {
					createNewRuleMap[helpers.CreateNonRootPortForwardingRuleName(port.Name, network.Name)] = cmp2
				}

				delete(ruleMap, helpers.CreateNonRootPortForwardingRuleName(port.Name, network.Name))
			}

		}

		// delete any remaining rules
		for _, rule := range ruleMap {
			err := client.DeletePortForwardingRule(rule.ID)
			if err != nil {
				return makeError(err)
			}
		}

		for name, public := range createNewRuleMap {
			// if the public port is 0 here, it means that request if for a new root rule
			if public.PublicPort == 0 {
				freePort, err := client.GetFreePort(zone.PortRange.Start, zone.PortRange.End)
				if err != nil {
					return makeError(err)
				}

				public.PublicPort = freePort
			}

			rule, err := helpers.CreatePortForwardingRule(client, vm, name, &public, vmModel.New().UpdateSubsystemByName)
			if err != nil {
				return makeError(err)
			}

			createNewRuleMap[name] = *rule
		}

		// merge the new and the old ones
		for name, public := range createNewRuleMap {
			newRuleMap[name] = public
		}

		err = vmModel.New().UpdateSubsystemByName(vm.Name, "cs", "portForwardingRuleMap", newRuleMap)
	}

	// service offering
	if !service.NotCreated(&vm.Subsystems.CS.ServiceOffering) {
		requiresUpdate := updateParams.CpuCores != nil && *updateParams.CpuCores != vm.Subsystems.CS.ServiceOffering.CpuCores ||
			updateParams.RAM != nil && *updateParams.RAM != vm.Subsystems.CS.ServiceOffering.RAM

		if requiresUpdate {
			public := helpers.CreateServiceOfferingPublic(vm.Name, *updateParams.CpuCores, *updateParams.RAM, vm.Specs.DiskSize)
			serviceOffering, err := helpers.RecreateServiceOffering(client, vm, public, vmModel.New().UpdateSubsystemByName)
			if err != nil {
				return makeError(err)
			}

			vm.Subsystems.CS.ServiceOffering = *serviceOffering
		}
	}

	// make sure the vm is using the latest service offering
	if vm.Subsystems.CS.VM.ServiceOfferingID != vm.Subsystems.CS.ServiceOffering.ID {
		vm.Subsystems.CS.VM.ServiceOfferingID = vm.Subsystems.CS.ServiceOffering.ID

		// turn it off if it is on, but remember the status
		status, err := client.GetVmStatus(vm.Subsystems.CS.VM.ID)
		if err != nil {
			return makeError(err)
		}

		if status == "Running" {
			err = client.DoVmCommand(vm.Subsystems.CS.VM.ID, nil, commands.Stop)
			if err != nil {
				return makeError(err)
			}
		}

		// update the service offering
		err = client.UpdateVM(&vm.Subsystems.CS.VM)
		if err != nil {
			return makeError(err)
		}

		err = vmModel.New().UpdateSubsystemByName(vm.Name, "cs", "vm.serviceOfferingId", vm.Subsystems.CS.ServiceOffering.ID)
		if err != nil {
			return makeError(err)
		}

		// turn it on if it was on
		if status == "Running" {
			var requiredHost *string
			if vm.HasGPU() {
				requiredHost, err = helpers.GetRequiredHost(vm.GpuID)
				if err != nil {
					return makeError(err)
				}
			}

			err = client.DoVmCommand(vm.Subsystems.CS.VM.ID, requiredHost, commands.Start)
			if err != nil {
				return makeError(err)
			}
		}

	}

	return nil
}

func RepairCS(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to repair cs %s. details: %w", name, err)
	}

	vm, err := vmModel.New().GetByName(name)
	if err != nil {
		return makeError(err)
	}

	if vm == nil {
		log.Println("vm", name, "not found when repairing cs, assuming it was deleted")
		return nil
	}

	zone := conf.Env.VM.GetZone(vm.Zone)
	if zone == nil {
		return makeError(fmt.Errorf("zone %s not found", vm.Zone))
	}

	client, err := helpers.WithCsClient(zone)
	if err != nil {
		return makeError(err)
	}

	ss := vm.Subsystems.CS

	// service offering
	err = helpers.RepairServiceOffering(client, vm, vmModel.New().UpdateSubsystemByName, func() *csModels.ServiceOfferingPublic {
		return helpers.CreateServiceOfferingPublic(name, vm.Specs.CpuCores, vm.Specs.RAM, vm.Specs.DiskSize)
	})

	// vm
	err = helpers.RepairVM(client, vm, vmModel.New().UpdateSubsystemByName, func() *csModels.VmPublic {
		if ss.ServiceOffering.Created() {
			return helpers.CreateCsVmPublic(name, ss.ServiceOffering.ID, TemplateID, helpers.CreateDeployTags(name, name))
		}
		return nil
	})
	if err != nil {
		return makeError(err)
	}

	// port-forwarding rules
	for mapName := range ss.GetPortForwardingRuleMap() {
		err = helpers.RepairPortForwardingRule(client, vm, mapName, storage_manager.UpdateSubsystemByID, func() *csModels.PortForwardingRulePublic {

			// find rules in vm ports
			for _, port := range vm.Ports {
				if port.Name == mapName {
					publicPort, err := client.GetFreePort(zone.PortRange.Start, zone.PortRange.End)
					if err != nil {
						log.Println("failed to get free port for port forwarding rule", mapName, "for vm", name, "in zone", zone.Name, ". details:", err)
						return nil
					}
					return helpers.CreateRootPortForwardingRulePublic(mapName, ss.VM.ID, publicPort, port.Port, port.Protocol, helpers.CreateDeployTags(mapName, ss.VM.Name), *zone)
				}
			}

			return nil
		})
	}

	return nil
}

func DoCommandCS(csVmID string, gpuID *string, command, zoneName string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to execute command %s for cs vm %s. details: %w", command, csVmID, err)
	}

	zone := conf.Env.VM.GetZone(zoneName)
	if zone == nil {
		return makeError(fmt.Errorf("zone %s not found", zoneName))
	}

	client, err := helpers.WithCsClient(zone)
	if err != nil {
		return makeError(err)
	}

	var requiredHost *string
	if gpuID != nil {
		requiredHost, err = helpers.GetRequiredHost(*gpuID)
		if err != nil {
			return makeError(err)
		}
	}

	err = client.DoVmCommand(csVmID, requiredHost, commands.Command(command))
	if err != nil {
		return makeError(err)
	}

	return nil
}

func CanStartCS(csVmID, hostName, zoneName string) (bool, string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if cs vm %s can be started on host %s. details: %w", csVmID, hostName, err)
	}

	zone := conf.Env.VM.GetZone(zoneName)
	if zone == nil {
		return false, "", makeError(fmt.Errorf("zone %s not found", zoneName))
	}

	client, err := helpers.WithCsClient(zone)
	if err != nil {
		return false, "", makeError(err)
	}

	hasCapacity, err := client.HasCapacity(csVmID, hostName)
	if err != nil {
		return false, "", err
	}

	if !hasCapacity {
		return false, "Host doesn't have capacity", nil
	}

	correctState, reason, err := HostInCorrectState(hostName, zone)
	if err != nil {
		return false, "", err
	}

	if !correctState {
		return false, reason, nil
	}

	return true, "", nil
}

func HostInCorrectState(hostName string, zone *enviroment.VmZone) (bool, string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if host %s is in correct state. details: %w", zone.Name, err)
	}

	client, err := helpers.WithCsClient(zone)
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
