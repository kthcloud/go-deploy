package helpers

import (
	"errors"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/pkg/subsystems/cs"
	csModels "go-deploy/pkg/subsystems/cs/models"
	"go-deploy/service"
)

func CreateServiceOffering(client *cs.Client, vm *vmModel.VM, public *csModels.ServiceOfferingPublic, updateDb service.UpdateDbSubsystem) (*csModels.ServiceOfferingPublic, error) {
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

	err = updateDb(vm.ID, "cs", "serviceOffering", serviceOffering)
	if err != nil {
		return nil, err
	}

	return serviceOffering, nil
}

func CreateCsVM(client *cs.Client, vm *vmModel.VM, public *csModels.VmPublic, userSshKey, adminSshKey string, updateDb service.UpdateDbSubsystem) (*csModels.VmPublic, error) {
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

	err = updateDb(vm.ID, "cs", "vm", csVM)
	if err != nil {
		return nil, err
	}

	vm.Subsystems.CS.VM = *csVM

	return csVM, nil
}

func CreatePortForwardingRule(client *cs.Client, vm *vmModel.VM, name string, public *csModels.PortForwardingRulePublic, updateDb service.UpdateDbSubsystem) (*csModels.PortForwardingRulePublic, error) {
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

	vm.Subsystems.CS.GetPortForwardingRuleMap()[name] = *portForwardingRule

	err = updateDb(vm.ID, "cs", "portForwardingRuleMap", vm.Subsystems.CS.GetPortForwardingRuleMap())
	if err != nil {
		return nil, err
	}

	return portForwardingRule, nil
}
