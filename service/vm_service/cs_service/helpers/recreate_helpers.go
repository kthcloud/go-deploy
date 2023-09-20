package helpers

import (
	csModels "go-deploy/pkg/subsystems/cs/models"
)

func (client *Client) RecreateServiceOffering(id string, public *csModels.ServiceOfferingPublic) (*csModels.ServiceOfferingPublic, error) {
	err := client.DeleteServiceOffering(client.CS.ServiceOffering.ID)
	if err != nil {
		return nil, err
	}

	return client.CreateServiceOffering(id, public)
}

func (client *Client) RecreatePortForwardingRule(id, name string, public *csModels.PortForwardingRulePublic) error {
	rule := client.CS.GetPortForwardingRule(name)
	if rule != nil {
		err := client.SsClient.DeletePortForwardingRule(rule.ID)
		if err != nil {
			return err
		}
	}

	_, err := client.CreatePortForwardingRule(id, name, public)
	if err != nil {
		return err
	}

	return nil
}
