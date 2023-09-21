package helpers

import (
	"go-deploy/pkg/conf"
	csModels "go-deploy/pkg/subsystems/cs/models"
	"go-deploy/service"
	"log"
	"strings"
)

func (client *Client) RepairServiceOffering(id string, genPublic func() *csModels.ServiceOfferingPublic) error {
	dbServiceOffering := &client.CS.ServiceOffering
	if service.NotCreated(dbServiceOffering) {
		public := genPublic()
		if public == nil {
			log.Println("no public supplied for service offering", id, "when trying to create it in the repair process")
			return nil
		}

		_, err := client.CreateServiceOffering(id, public)
		return err
	}

	return service.UpdateIfDiff(
		*dbServiceOffering,
		func() (*csModels.ServiceOfferingPublic, error) {
			return client.SsClient.ReadServiceOffering(dbServiceOffering.ID)
		},
		client.SsClient.UpdateServiceOffering,
		func(serviceOffering *csModels.ServiceOfferingPublic) error {
			_, err := client.SsClient.CreateServiceOffering(serviceOffering)
			return err
		},
	)
}

func (client *Client) RepairVM(id string, genPublic func() (*csModels.VmPublic, string)) error {
	dbVM := &client.CS.VM
	if service.NotCreated(dbVM) {
		adminSshKey := conf.Env.VM.AdminSshPublicKey
		public, userSshKey := genPublic()
		if public == nil {
			log.Println("no public supplied for vm", id, "when trying to create it in the repair process")
			return nil
		}

		_, err := client.CreateCsVM(id, public, userSshKey, adminSshKey)
		return err
	}

	// we can't just recreate a virtual machine, so we only bother trying to update it

	return service.UpdateIfDiff(
		*dbVM,
		func() (*csModels.VmPublic, error) {
			return client.SsClient.ReadVM(dbVM.ID)
		},
		client.SsClient.UpdateVM,
		func(vm *csModels.VmPublic) error { return nil },
	)
}

func (client *Client) RepairPortForwardingRule(id, name string, genPublic func() *csModels.PortForwardingRulePublic) error {
	rule := client.CS.GetPortForwardingRule(name)
	if service.NotCreated(rule) {
		public := genPublic()
		if public == nil {
			log.Println("no public supplied for port forwarding rule", name, " when trying to create it in the repair process")
			return nil
		}

		_, err := client.CreatePortForwardingRule(id, name, public)
		if err != nil {
			if strings.Contains(err.Error(), "conflicts with rule") {
				// if we fail here it might be because the port was snatched
				// by another vm, so we try to recreate the rule with a new port

				freePort, withNewPortErr := client.GetFreePort()
				if withNewPortErr != nil {
					return withNewPortErr
				}

				oldPort := public.PublicPort
				public.PublicPort = freePort

				_, withNewPortErr = client.CreatePortForwardingRule(id, name, public)
				if withNewPortErr != nil {
					public.PublicPort = oldPort
					return withNewPortErr
				}

				return nil
			}
		}

		return err
	}

	return service.UpdateIfDiff(
		*rule,
		func() (*csModels.PortForwardingRulePublic, error) {
			return client.SsClient.ReadPortForwardingRule(rule.ID)
		},
		client.SsClient.UpdatePortForwardingRule,
		func(rule *csModels.PortForwardingRulePublic) error {
			return client.RecreatePortForwardingRule(id, name, rule)
		},
	)
}
