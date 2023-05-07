package internal_service

import (
	"fmt"
	vmModel "go-deploy/models/sys/vm"
	gpu2 "go-deploy/models/sys/vm/gpu"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/cs"
	"go-deploy/pkg/subsystems/cs/commands"
	csModels "go-deploy/pkg/subsystems/cs/models"
	"log"
	"strings"
)

type CsCreated struct {
	VM                 *csModels.VmPublic
	PortForwardingRule *csModels.PortForwardingRulePublic
	PublicIpAddress    *csModels.PublicIpAddressPublic
}

func withClient() (*cs.Client, error) {
	return cs.New(&cs.ClientConf{
		URL:       conf.Env.CS.URL,
		ApiKey:    conf.Env.CS.ApiKey,
		Secret:    conf.Env.CS.Secret,
		NetworkID: conf.Env.CS.NetworkID,
		ProjectID: conf.Env.CS.ProjectID,
		ZoneID:    conf.Env.CS.ZoneID,
	})
}

func CreateCS(name, sshPublicKey string, ports []vmModel.Port) (*CsCreated, error) {
	log.Println("setting up cs for", name)

	if sshPublicKey == "" {
		return nil, fmt.Errorf("ssh public key is required")
	}

	userSshPublicKey := sshPublicKey
	adminSshPublicKey := conf.Env.VM.AdminSshPublicKey

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup cs for vm %s. details: %s", name, err)
	}

	client, err := withClient()
	if err != nil {
		return nil, makeError(err)
	}

	vm, err := vmModel.GetByName(name)
	if err != nil {
		return nil, makeError(err)
	}

	if vm == nil {
		// if vm does not exist, don't treat as error, don't create -> job will not fail
		return nil, nil
	}

	// vm
	var csVM *csModels.VmPublic
	if vm.Subsystems.CS.VM.ID == "" {
		id, err := client.CreateVM(&csModels.VmPublic{
			Name: name,
			// temporary until vm templates are set up
			ServiceOfferingID: "8da28b4d-5fec-4a44-aee7-fb0c5c8265a9", // Small HA
			TemplateID:        "fb6b6b11-6196-42d9-a12d-038bdeecb6f6", // deploy-template-cloud-init-ubuntu2204
			Tags: []csModels.Tag{
				{Key: "name", Value: name},
				{Key: "managedBy", Value: conf.Env.Manager},
				{Key: "deployName", Value: name},
			},
		},
			userSshPublicKey, adminSshPublicKey,
		)
		if err != nil {
			return nil, makeError(err)
		}

		csVM, err = client.ReadVM(id)
		if err != nil {
			return nil, makeError(err)
		}

		err = vmModel.UpdateSubsystemByName(name, "cs", "vm", *csVM)
		if err != nil {
			return nil, makeError(err)
		}
	} else {
		csVM = &vm.Subsystems.CS.VM
	}

	// ip address
	var publicIpAddress *csModels.PublicIpAddressPublic
	if vm.Subsystems.CS.PublicIpAddress.ID == "" {
		public := &csModels.PublicIpAddressPublic{
			Name: name,
			Tags: []csModels.Tag{
				{Key: "name", Value: "main"},
				{Key: "managedBy", Value: conf.Env.Manager},
				{Key: "deployName", Value: name},
			},
		}

		id, err := client.CreatePublicIpAddress(public)
		if err != nil {
			return nil, makeError(err)
		}

		publicIpAddress, err = client.ReadPublicIpAddress(id)
		if err != nil {
			return nil, makeError(err)
		}

		err = vmModel.UpdateSubsystemByName(name, "cs", "publicIpAddress", *publicIpAddress)
		if err != nil {
			return nil, makeError(err)
		}
	} else {
		publicIpAddress = &vm.Subsystems.CS.PublicIpAddress
	}

	// port-forwarding rule
	var portForwardingRule *csModels.PortForwardingRulePublic

	addDeploySshToPortMap(&ports)

	ruleMap := vm.Subsystems.CS.PortForwardingRuleMap
	if ruleMap == nil {
		ruleMap = map[string]csModels.PortForwardingRulePublic{}
	}

	for _, port := range ports {
		rule, hasRule := ruleMap[port.Name]
		if hasRule && rule.ID != "" {
			// make sure this is connected to the ip address above by deleting the one we think we own
			if rule.IpAddressID != publicIpAddress.ID || rule.VmID != csVM.ID {
				err = client.DeletePortForwardingRule(rule.ID)
				if err != nil {
					return nil, makeError(err)
				}

				delete(ruleMap, port.Name)

				err = vmModel.UpdateSubsystemByName(name, "cs", "portForwardingRuleMap", ruleMap)
				if err != nil {
					return nil, makeError(err)
				}
			}
		}

		rule, hasRule = ruleMap[port.Name]
		if !hasRule || rule.ID == "" {
			id, err := client.CreatePortForwardingRule(&csModels.PortForwardingRulePublic{
				VmID:        csVM.ID,
				Name:        name,
				IpAddressID: publicIpAddress.ID,
				Protocol:    port.Protocol,
				PublicPort:  port.Port,
				PrivatePort: port.Port,
				Tags: []csModels.Tag{
					{Key: "name", Value: port.Name},
					{Key: "managedBy", Value: conf.Env.Manager},
					{Key: "deployName", Value: name},
				},
			})
			if err != nil {
				return nil, makeError(err)
			}

			rule, err := client.ReadPortForwardingRule(id)
			if err != nil {
				return nil, makeError(err)
			}

			if ruleMap == nil {
				return nil, makeError(fmt.Errorf("failed to read port forwarding rule after creation"))
			}

			ruleMap[port.Name] = *rule

			err = vmModel.UpdateSubsystemByName(name, "cs", "portForwardingRuleMap", ruleMap)
			if err != nil {
				return nil, makeError(err)
			}
		}
	}

	return &CsCreated{
		VM:                 csVM,
		PortForwardingRule: portForwardingRule,
		PublicIpAddress:    publicIpAddress,
	}, nil
}

func DeleteCS(name string) error {
	log.Println("deleting cs for", name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to delete cs for vm %s. details: %s", name, err)
	}

	client, err := withClient()
	if err != nil {
		return makeError(err)
	}

	vm, err := vmModel.GetByName(name)
	if err != nil {
		return makeError(err)
	}

	if vm == nil {
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

	if vm.Subsystems.CS.PublicIpAddress.ID != "" {
		err = client.DeletePublicIpAddress(vm.Subsystems.CS.PublicIpAddress.ID)
		if err != nil {
			return makeError(err)
		}

		err = vmModel.UpdateSubsystemByName(name, "cs", "publicIpAddress", csModels.PublicIpAddressPublic{})
		if err != nil {
			return makeError(err)
		}
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

	return nil
}

func UpdateCS(vmID string, ports *[]vmModel.Port) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update cs for vm %s. details: %s", vmID, err)
	}

	client, err := withClient()
	if err != nil {
		return makeError(err)
	}

	vm, err := vmModel.GetByID(vmID)
	if err != nil {
		return makeError(err)
	}

	if vm == nil {
		return nil
	}

	if vm.Subsystems.CS.VM.ID == "" {
		return makeError(fmt.Errorf("vm is not created yet"))
	}

	if vm.Subsystems.CS.PublicIpAddress.ID == "" {
		return makeError(fmt.Errorf("public ip address is not created yet"))
	}

	// port-forwarding rule
	if ports == nil {
		return nil
	}

	removeDeploySshFromPortMap(ports)

	/// delete old rules and create new ones
	ruleMap := vm.Subsystems.CS.PortForwardingRuleMap

	currentPortForwardingRules, err := client.ReadPortForwardingRules(vm.Subsystems.CS.VM.ID)
	if err != nil {
		return makeError(err)
	}

	currentPorts := convertToPorts(currentPortForwardingRules)
	for i, port := range currentPorts {
		if port.Name == "__ssh" || port.Port == 22 {
			continue
		}

		err = client.DeletePortForwardingRule(currentPortForwardingRules[i].ID)
		if err != nil {
			return makeError(err)
		}

		delete(ruleMap, port.Name)

		err = vmModel.UpdateSubsystemByName(vm.Name, "cs", "portForwardingRuleMap", ruleMap)
		if err != nil {
			return makeError(err)
		}
	}

	for _, port := range *ports {
		id, err := client.CreatePortForwardingRule(&csModels.PortForwardingRulePublic{
			Name:        port.Name,
			VmID:        vm.Subsystems.CS.VM.ID,
			IpAddressID: vm.Subsystems.CS.PublicIpAddress.ID,
			Protocol:    port.Protocol,
			PublicPort:  port.Port,
			PrivatePort: port.Port,
			Tags: []csModels.Tag{
				{Key: "name", Value: port.Name},
				{Key: "managedBy", Value: conf.Env.Manager},
				{Key: "deployName", Value: vm.Name},
			},
		})
		if err != nil {
			return makeError(err)
		}

		rule, err := client.ReadPortForwardingRule(id)
		if err != nil {
			return makeError(err)
		}

		ruleMap[port.Name] = *rule

		err = vmModel.UpdateSubsystemByName(vm.Name, "cs", "portForwardingRuleMap", ruleMap)
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

func AttachGPU(gpuID, vmID string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to attach gpu %s to cs vm %s. details: %s", gpuID, vmID, err)
	}

	client, err := withClient()
	if err != nil {
		return makeError(err)
	}

	vm, err := vmModel.GetByID(vmID)
	if err != nil {
		return makeError(err)
	}

	if vm == nil {
		return makeError(fmt.Errorf("vm %s not found", vmID))
	}

	if vm.Subsystems.CS.VM.ID == "" {
		return makeError(fmt.Errorf("vm is not created yet"))
	}

	gpu, err := gpu2.GetGpuByID(gpuID)
	if err != nil {
		return makeError(err)
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

	vm.Subsystems.CS.VM.ExtraConfig = createExtraConfig(gpu)

	err = client.UpdateVM(&vm.Subsystems.CS.VM)
	if err != nil {
		return makeError(err)
	}

	err = vmModel.UpdateSubsystemByName(vm.Name, "cs", "vm.extraConfig", vm.Subsystems.CS.VM.ExtraConfig)
	if err != nil {
		return makeError(err)
	}

	requiredHost, err := getRequiredHost(gpuID)
	if err != nil {
		return makeError(err)
	}

	// turn it on if it was on
	if status == "Running" {
		err = client.DoVmCommand(vm.Subsystems.CS.VM.ID, requiredHost, "start")
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

func DetachGPU(vmID string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to detach gpu from cs vm %s. details: %s", vmID, err)
	}

	client, err := withClient()
	if err != nil {
		return makeError(err)
	}

	vm, err := vmModel.GetByID(vmID)
	if err != nil {
		return makeError(err)
	}

	if vm == nil {
		return makeError(fmt.Errorf("vm %s not found", vmID))
	}

	if vm.Subsystems.CS.VM.ID == "" {
		return makeError(fmt.Errorf("vm is not created yet"))
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

	client, err := withClient()
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

func DoCommandCS(vmID string, gpuID *string, command string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to execute command %s for cs vm %s. details: %s", command, vmID, err)
	}

	client, err := withClient()
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

func createExtraConfig(gpu *gpu2.GPU) string {
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
	gpu, err := gpu2.GetGpuByID(gpuID)
	if err != nil {
		return nil, err
	}

	if gpu.Host == "" {
		return nil, fmt.Errorf("no host found for gpu %s", gpu.ID)
	}

	return &gpu.Host, nil
}

func addDeploySshToPortMap(portMap *[]vmModel.Port) {
	for i, port := range *portMap {
		if (port.Port == 22 || port.Name == "__ssh") && port.Protocol == "tcp" {
			*portMap = append((*portMap)[:i], (*portMap)[i+1:]...)
			break
		}
	}

	*portMap = append(*portMap, vmModel.Port{
		Port:     22,
		Name:     "__ssh",
		Protocol: "tcp",
	})
}

func removeDeploySshFromPortMap(portMap *[]vmModel.Port) {
	for i, port := range *portMap {
		if (port.Port == 22 || port.Name == "__ssh") && port.Protocol == "tcp" {
			*portMap = append((*portMap)[:i], (*portMap)[i+1:]...)
			break
		}
	}
}

func convertToPorts(rules []csModels.PortForwardingRulePublic) []vmModel.Port {
	var ports []vmModel.Port

	for _, rule := range rules {
		var name string
		for _, tag := range rule.Tags {
			if tag.Key == "name" {
				name = tag.Value
				break
			}
		}

		ports = append(ports, vmModel.Port{
			Port:     rule.PublicPort,
			Name:     name,
			Protocol: rule.Protocol,
		})
	}

	return ports
}
