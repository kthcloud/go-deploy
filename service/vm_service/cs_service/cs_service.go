package cs_service

import (
	"fmt"
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

	vm, err := vmModel.New().GetByName(params.Name)
	if err != nil {
		return nil, makeError(err)
	}

	client, err := helpers.New(&vm.Subsystems.CS, params.Zone)
	if err != nil {
		return nil, makeError(err)
	}

	if vm == nil {
		log.Println("vm", params.Name, "not found for cs setup. assuming it was deleted")
		return nil, nil
	}

	// service offering
	serviceOffering := &client.CS.ServiceOffering
	if !serviceOffering.Created() {
		public := helpers.CreateServiceOfferingPublic(params.Name, params.CpuCores, params.RAM, params.DiskSize)
		serviceOffering, err = client.CreateServiceOffering(vm.ID, public)
		if err != nil {
			return nil, makeError(err)
		}
	}

	// vm
	csVM := &client.CS.VM
	if !csVM.Created() {
		public := helpers.CreateCsVmPublic(params.Name, serviceOffering.ID, TemplateID, CreateDeployTags(params.Name, params.Name))

		csVM, err = client.CreateCsVM(vm.ID, public, userSshPublicKey, adminSshPublicKey)
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

	ruleMap := client.CS.GetPortForwardingRuleMap()
	if ruleMap == nil {
		ruleMap = map[string]csModels.PortForwardingRulePublic{}
	}

	for _, port := range params.Ports {
		if rule, hasRule := ruleMap[port.Name]; hasRule && rule.Created() {
			continue
		}

		freePort, err := client.GetFreePort()
		if err != nil {
			return nil, makeError(err)
		}

		useRootNetwork := params.NetworkID == nil
		if useRootNetwork {
			rootPublic := helpers.CreateRootPortForwardingRulePublic(port.Name, csVM.ID, freePort, port.Port, port.Protocol, CreateDeployTags(port.Name, csVM.Name), *client.Zone)
			_, err = client.CreatePortForwardingRule(vm.ID, port.Name, rootPublic)
			if err != nil {
				return nil, makeError(err)
			}
		} else {
			network, err := client.SsClient.ReadNetwork(*params.NetworkID)
			if err != nil {
				return nil, makeError(err)
			}

			rootPublic := helpers.CreateRootPortForwardingRulePublic(
				port.Name,
				csVM.ID,
				freePort,
				port.Port,
				port.Protocol,
				CreateDeployTags(port.Name, csVM.Name),
				*client.Zone,
			)

			_, err = client.CreatePortForwardingRule(vm.ID, port.Name, rootPublic)
			if err != nil {
				return nil, makeError(err)
			}

			ipAddressID, err := client.SsClient.GetNetworkSourceNatIpAddressID(*params.NetworkID)
			if err != nil {
				return nil, makeError(err)
			}

			public := helpers.CreatePortForwardingRulePublic(
				CreateNonRootPortForwardingRuleName(port.Name, network.Name),
				csVM.ID,
				*params.NetworkID,
				ipAddressID,
				port.Port,
				port.Protocol,
				CreateDeployTags(port.Name, csVM.Name),
			)
			_, err = client.CreatePortForwardingRule(vm.ID, port.Name, public)
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

	client, err := helpers.New(&vm.Subsystems.CS, vm.Zone)
	if err != nil {
		return makeError(err)
	}

	for mapName := range client.CS.GetPortForwardingRuleMap() {
		err = client.DeletePortForwardingRule(vm.ID, mapName)
		if err != nil {
			return makeError(err)
		}
	}

	err = client.DeleteVM(vm.ID)
	if err != nil {
		return makeError(err)
	}

	err = client.DeleteServiceOffering(vm.ID)
	if err != nil {
		return makeError(err)
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

	client, err := helpers.New(&vm.Subsystems.CS, vm.Zone)
	if err != nil {
		return makeError(err)
	}

	if !client.CS.VM.Created() {
		log.Println("cs vm", vmID, "not created when updating cs. assuming it was deleted or not yet created")
		return nil
	}

	// port-forwarding rule
	if updateParams.Ports != nil {
		ruleMap := client.CS.GetPortForwardingRuleMap()

		newRuleMap := make(map[string]csModels.PortForwardingRulePublic)
		createNewRuleMap := make(map[string]csModels.PortForwardingRulePublic)

		for _, port := range *updateParams.Ports {
			useRootNetwork := vm.NetworkID == nil
			cmp1 := helpers.CreateRootPortForwardingRulePublic(
				port.Name,
				client.CS.VM.ID,
				// set this to zero here, in case we don't need a new rule (getting a free port is expensive)
				0, port.Port,
				port.Protocol,
				CreateDeployTags(port.Name, vm.Name),
				*client.Zone,
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
				err = client.SsClient.DeletePortForwardingRule(cmp2.ID)
				if err != nil {
					return makeError(err)
				}

				createNewRuleMap[port.Name] = *cmp1
			} else {
				createNewRuleMap[port.Name] = cmp2
			}

			delete(ruleMap, port.Name)

			if !useRootNetwork {
				network, err := client.SsClient.ReadNetwork(*vm.NetworkID)
				if err != nil {
					return makeError(err)
				}

				ruleName := CreateNonRootPortForwardingRuleName(port.Name, network.Name)

				cmp1 = helpers.CreatePortForwardingRulePublic(
					ruleName,
					client.CS.VM.ID,
					*vm.NetworkID,
					cmp2.IpAddressID,
					port.Port,
					port.Protocol,
					CreateDeployTags(port.Name, vm.Name),
				)
				cmp2, ok = ruleMap[ruleName]
				if !ok {
					// new rule
					createNewRuleMap[ruleName] = *cmp1
					delete(ruleMap, ruleName)
					continue
				}

				cmp2Cleaned = cmp2
				cmp2Cleaned.ID = ""
				cmp2Cleaned.CreatedAt = time.Time{}

				if !reflect.DeepEqual(cmp1, cmp2Cleaned) {
					err = client.SsClient.DeletePortForwardingRule(cmp2.ID)
					if err != nil {
						return makeError(err)
					}

					createNewRuleMap[CreateNonRootPortForwardingRuleName(port.Name, network.Name)] = *cmp1
				} else {
					createNewRuleMap[CreateNonRootPortForwardingRuleName(port.Name, network.Name)] = cmp2
				}

				delete(ruleMap, CreateNonRootPortForwardingRuleName(port.Name, network.Name))
			}

		}

		// delete any remaining rules
		for _, rule := range ruleMap {
			err := client.SsClient.DeletePortForwardingRule(rule.ID)
			if err != nil {
				return makeError(err)
			}
		}

		for name, public := range createNewRuleMap {
			// if the public port is 0 here, it means that request if for a new root rule
			if public.PublicPort == 0 {
				freePort, err := client.GetFreePort()
				if err != nil {
					return makeError(err)
				}

				public.PublicPort = freePort
			}

			rule, err := client.CreatePortForwardingRule(vm.ID, name, &public)
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
	if !service.NotCreated(&client.CS.ServiceOffering) {
		requiresUpdate := updateParams.CpuCores != nil && *updateParams.CpuCores != client.CS.ServiceOffering.CpuCores ||
			updateParams.RAM != nil && *updateParams.RAM != client.CS.ServiceOffering.RAM

		if requiresUpdate {
			public := helpers.CreateServiceOfferingPublic(vm.Name, *updateParams.CpuCores, *updateParams.RAM, vm.Specs.DiskSize)
			serviceOffering, err := client.RecreateServiceOffering(vm.ID, public)
			if err != nil {
				return makeError(err)
			}

			client.CS.ServiceOffering = *serviceOffering
		}
	}

	// make sure the vm is using the latest service offering
	if client.CS.VM.ServiceOfferingID != client.CS.ServiceOffering.ID {
		client.CS.VM.ServiceOfferingID = client.CS.ServiceOffering.ID

		// turn it off if it is on, but remember the status
		status, err := client.SsClient.GetVmStatus(client.CS.VM.ID)
		if err != nil {
			return makeError(err)
		}

		if status == "Running" {
			err = client.SsClient.DoVmCommand(client.CS.VM.ID, nil, commands.Stop)
			if err != nil {
				return makeError(err)
			}
		}

		// update the service offering
		err = client.SsClient.UpdateVM(&client.CS.VM)
		if err != nil {
			return makeError(err)
		}

		err = vmModel.New().UpdateSubsystemByName(vm.Name, "cs", "vm.serviceOfferingId", client.CS.ServiceOffering.ID)
		if err != nil {
			return makeError(err)
		}

		// turn it on if it was on
		if status == "Running" {
			var requiredHost *string
			if vm.HasGPU() {
				requiredHost, err = GetRequiredHost(vm.GpuID)
				if err != nil {
					return makeError(err)
				}
			}

			err = client.SsClient.DoVmCommand(client.CS.VM.ID, requiredHost, commands.Start)
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

	client, err := helpers.New(&vm.Subsystems.CS, vm.Zone)
	if err != nil {
		return makeError(err)
	}

	ss := client.CS

	// service offering
	err = client.RepairServiceOffering(vm.ID, func() *csModels.ServiceOfferingPublic {
		return helpers.CreateServiceOfferingPublic(name, vm.Specs.CpuCores, vm.Specs.RAM, vm.Specs.DiskSize)
	})

	// vm
	err = client.RepairVM(vm.ID, func() (*csModels.VmPublic, string) {
		if ss.ServiceOffering.Created() {
			return helpers.CreateCsVmPublic(name, ss.ServiceOffering.ID, TemplateID, CreateDeployTags(name, name)), vm.SshPublicKey
		}
		return nil, ""
	})
	if err != nil {
		return makeError(err)
	}

	// port-forwarding rules
	for mapName := range ss.GetPortForwardingRuleMap() {
		err = client.RepairPortForwardingRule(vm.ID, mapName, func() *csModels.PortForwardingRulePublic {
			// find rules in vm ports
			for _, port := range vm.Ports {
				if port.Name == mapName {
					publicPort, err := client.GetFreePort()
					if err != nil {
						log.Println("failed to get free port for port forwarding rule", mapName, "for vm", name, "in zone", client.Zone.Name, ". details:", err)
						return nil
					}
					return helpers.CreateRootPortForwardingRulePublic(mapName, ss.VM.ID, publicPort, port.Port, port.Protocol, CreateDeployTags(mapName, ss.VM.Name), *client.Zone)
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

	client, err := helpers.New(nil, zoneName)
	if err != nil {
		return makeError(err)
	}

	var requiredHost *string
	if gpuID != nil {
		requiredHost, err = GetRequiredHost(*gpuID)
		if err != nil {
			return makeError(err)
		}
	}

	err = client.SsClient.DoVmCommand(csVmID, requiredHost, commands.Command(command))
	if err != nil {
		return makeError(err)
	}

	return nil
}

func CanStartCS(csVmID, hostName, zoneName string) (bool, string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if cs vm %s can be started on host %s. details: %w", csVmID, hostName, err)
	}

	client, err := helpers.New(nil, zoneName)
	if err != nil {
		return false, "", makeError(err)
	}

	hasCapacity, err := client.SsClient.HasCapacity(csVmID, hostName)
	if err != nil {
		return false, "", err
	}

	if !hasCapacity {
		return false, "Host doesn't have capacity", nil
	}

	correctState, reason, err := HostInCorrectState(hostName, client.Zone)
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

	client, err := helpers.New(nil, zone.Name)
	if err != nil {
		return false, "", makeError(err)
	}

	host, err := client.SsClient.ReadHostByName(hostName)
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
