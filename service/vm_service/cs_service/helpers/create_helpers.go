package helpers

import (
	"errors"
	csModels "go-deploy/pkg/subsystems/cs/models"
)

func (client *Client) CreateServiceOffering(id string, public *csModels.ServiceOfferingPublic) (*csModels.ServiceOfferingPublic, error) {
	createdID, err := client.SsClient.CreateServiceOffering(public)
	if err != nil {
		return nil, err
	}

	serviceOffering, err := client.SsClient.ReadServiceOffering(createdID)
	if err != nil {
		return nil, err
	}

	if serviceOffering == nil {
		return nil, errors.New("failed to read service offering after creation")
	}

	err = client.UpdateDB(id, "serviceOffering", serviceOffering)
	if err != nil {
		return nil, err
	}

	client.CS.ServiceOffering = *serviceOffering

	return serviceOffering, nil
}

func (client *Client) CreateCsVM(id string, public *csModels.VmPublic, userSshKey, adminSshKey string) (*csModels.VmPublic, error) {
	createdID, err := client.SsClient.CreateVM(public, userSshKey, adminSshKey)
	if err != nil {
		return nil, err
	}

	csVM, err := client.SsClient.ReadVM(createdID)
	if err != nil {
		return nil, err
	}

	if csVM == nil {
		return nil, errors.New("failed to read vm after creation")
	}

	err = client.UpdateDB(id, "vm", csVM)
	if err != nil {
		return nil, err
	}

	client.CS.VM = *csVM

	return csVM, nil
}

func (client *Client) CreatePortForwardingRule(id, name string, public *csModels.PortForwardingRulePublic) (*csModels.PortForwardingRulePublic, error) {
	createdID, err := client.SsClient.CreatePortForwardingRule(public)
	if err != nil {
		return nil, err
	}

	portForwardingRule, err := client.SsClient.ReadPortForwardingRule(createdID)
	if err != nil {
		return nil, err
	}

	if portForwardingRule == nil {
		return nil, errors.New("failed to read port forwarding rule after creation")
	}

	err = client.UpdateDB(id, "portForwardingRuleMap."+name, *portForwardingRule)
	if err != nil {
		return nil, err
	}

	client.CS.SetPortForwardingRule(name, *portForwardingRule)

	return portForwardingRule, nil
}
