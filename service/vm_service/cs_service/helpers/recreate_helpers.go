package helpers

import (
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/pkg/subsystems/cs"
	csModels "go-deploy/pkg/subsystems/cs/models"
	"go-deploy/service"
)

func RecreateServiceOffering(client *cs.Client, vm *vmModel.VM, public *csModels.ServiceOfferingPublic, updateDb service.UpdateDbSubsystem) (*csModels.ServiceOfferingPublic, error) {
	err := client.DeleteServiceOffering(vm.Subsystems.CS.ServiceOffering.ID)
	if err != nil {
		return nil, err
	}

	return CreateServiceOffering(client, vm, public, updateDb)
}

func RecreatePortForwardingRule(client *cs.Client, vm *vmModel.VM, name string, public *csModels.PortForwardingRulePublic, updateDb service.UpdateDbSubsystem) error {
	rule, ok := vm.Subsystems.CS.GetPortForwardingRuleMap()[name]
	if ok {
		err := client.DeletePortForwardingRule(rule.ID)
		if err != nil {
			return err
		}
	}

	_, err := CreatePortForwardingRule(client, vm, name, public, updateDb)
	if err != nil {
		return err
	}

	return nil
}
