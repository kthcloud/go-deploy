package helpers

import (
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/cs"
	csModels "go-deploy/pkg/subsystems/cs/models"
	"go-deploy/service"
	"log"
)

func RepairServiceOffering(client *cs.Client, vm *vmModel.VM, updateDb service.UpdateDbSubsystem, genPublic func() *csModels.ServiceOfferingPublic) error {
	dbServiceOffering := &vm.Subsystems.CS.ServiceOffering
	if service.NotCreated(dbServiceOffering) {
		public := genPublic()
		if public == nil {
			log.Println("no public supplied for service offering when trying to create it in the repair process")
			return nil
		}

		_, err := CreateServiceOffering(client, vm, public, updateDb)
		return err
	}

	return service.UpdateIfDiff(
		*dbServiceOffering,
		func() (*csModels.ServiceOfferingPublic, error) {
			return client.ReadServiceOffering(dbServiceOffering.ID)
		},
		client.UpdateServiceOffering,
		func(serviceOffering *csModels.ServiceOfferingPublic) error {
			_, err := CreateServiceOffering(client, vm, serviceOffering, updateDb)
			return err
		},
	)
}

func RepairVM(client *cs.Client, vm *vmModel.VM, updateDb service.UpdateDbSubsystem, genPublic func() *csModels.VmPublic) error {
	dbVM := &vm.Subsystems.CS.VM
	if service.NotCreated(dbVM) {
		adminSshKey := conf.Env.VM.AdminSshPublicKey
		public := genPublic()
		if public == nil {
			log.Println("no public supplied for vm", vm.ID, " when trying to create it in the repair process")
			return nil
		}

		_, err := CreateCsVM(client, vm, public, vm.SshPublicKey, adminSshKey, updateDb)
		return err
	}

	// we can't just recreate a virtual machine, so we only bother trying to update it

	return service.UpdateIfDiff(
		*dbVM,
		func() (*csModels.VmPublic, error) {
			return client.ReadVM(dbVM.ID)
		},
		client.UpdateVM,
		func(vm *csModels.VmPublic) error { return nil },
	)
}

func RepairPortForwardingRule(client *cs.Client, vm *vmModel.VM, name string, updateDb service.UpdateDbSubsystem, genPublic func() *csModels.PortForwardingRulePublic) error {
	rule := vm.Subsystems.CS.GetPortForwardingRule(name)
	if service.NotCreated(rule) {
		public := genPublic()
		if public == nil {
			log.Println("no public supplied for port forwarding rule", name, " when trying to create it in the repair process")
			return nil
		}

		_, err := CreatePortForwardingRule(client, vm, name, public, updateDb)
		return err
	}

	return service.UpdateIfDiff(
		*rule,
		func() (*csModels.PortForwardingRulePublic, error) {
			return client.ReadPortForwardingRule(rule.ID)
		},
		client.UpdatePortForwardingRule,
		func(rule *csModels.PortForwardingRulePublic) error {
			return RecreatePortForwardingRule(client, vm, name, rule, updateDb)
		},
	)
}
