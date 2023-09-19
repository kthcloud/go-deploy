package helpers

import (
	csModels "go-deploy/pkg/subsystems/cs/models"
	"go-deploy/service"
)

func (client *Client) DeleteServiceOffering(id string) error {
	if client.CS.ServiceOffering.Created() {
		err := client.SsClient.DeleteServiceOffering(client.CS.ServiceOffering.ID)
		if err != nil {
			return err
		}

		err = client.UpdateDB(id, "serviceOffering", nil)

		client.CS.ServiceOffering = csModels.ServiceOfferingPublic{}
	}

	return nil
}

func (client *Client) DeleteVM(id string) error {
	if client.CS.VM.Created() {
		err := client.SsClient.DeleteVM(client.CS.VM.ID)
		if err != nil {
			return err
		}

		err = client.UpdateDB(id, "vm", nil)
		if err != nil {
			return err
		}

		client.CS.VM = csModels.VmPublic{}
	}

	return nil
}

func (client *Client) DeletePortForwardingRule(id, name string) error {
	rule := client.CS.GetPortForwardingRule(name)

	if service.Created(rule) {
		err := client.SsClient.DeletePortForwardingRule(rule.ID)
		if err != nil {
			return err
		}

		client.CS.DeletePortForwardingRule(name)

		err = client.UpdateDB(id, "portForwardingRuleMap", client.CS.PortForwardingRuleMap)
		if err != nil {
			return err
		}
	}

	return nil
}
