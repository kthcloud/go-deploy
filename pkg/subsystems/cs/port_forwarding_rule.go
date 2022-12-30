package cs

import (
	"errors"
	"fmt"
	"go-deploy/pkg/subsystems/cs/models"
	"strings"
)

func (client *Client) ReadPortForwardingRule(id string) (*models.PortForwardingRulePublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read port forwarding rule %s. details: %s", id, err)
	}

	if id == "" {
		return nil, fmt.Errorf("id required")
	}

	portForwardingRule, _, err := client.CSClient.Firewall.GetPortForwardingRuleByID(id)
	if err != nil {
		if !strings.Contains(err.Error(), "No match found for") {
			return nil, makeError(err)
		}
	}

	// fetch project id
	vm, _, err := client.CSClient.VirtualMachine.GetVirtualMachineByID(portForwardingRule.Virtualmachineid)
	if err != nil {
		return nil, makeError(err)
	}

	var public *models.PortForwardingRulePublic
	if portForwardingRule != nil {
		public = models.CreatePortForwardingRulePublicFromGet(portForwardingRule, vm.Projectid)
	}

	return public, nil
}

func (client *Client) CreatePortForwardingRule(public *models.PortForwardingRulePublic) (string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create port forwarding rule for vm %s. details: %s", public.VmID, err)
	}

	if public.VmID == "" {
		return "", makeError(errors.New("vm id required"))
	}

	if public.NetworkID == "" {
		return "", makeError(errors.New("network id required"))
	}

	if public.IpAddressID == "" {
		return "", makeError(errors.New("ip address id required"))
	}

	listRulesParams := client.CSClient.Firewall.NewListPortForwardingRulesParams()
	listRulesParams.SetProjectid(public.ProjectID)
	listRulesParams.SetNetworkid(public.NetworkID)
	listRulesParams.SetIpaddressid(public.IpAddressID)
	listRulesParams.SetListall(true)

	portForwardingRulesResponse, err := client.CSClient.Firewall.ListPortForwardingRules(listRulesParams)
	if err != nil {
		return "", makeError(err)
	}

	if portForwardingRulesResponse.Count != 0 {
		for _, rule := range portForwardingRulesResponse.PortForwardingRules {
			if rule.Virtualmachineid == public.VmID {
				return rule.Id, nil
			}
		}
	}

	createRuleParams := client.CSClient.Firewall.NewCreatePortForwardingRuleParams(
		public.IpAddressID,
		public.PrivatePort,
		public.Protocol,
		public.PublicPort,
		public.VmID,
	)

	created, err := client.CSClient.Firewall.CreatePortForwardingRule(createRuleParams)
	if err != nil {
		return "", makeError(err)
	}

	return created.Id, nil
}

func (client *Client) DeletePortForwardingRule(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete port forwarding rule %s. details: %s", id, err)
	}

	if id == "" {
		return fmt.Errorf("id required")
	}

	portForwardingRule, _, err := client.CSClient.Firewall.GetPortForwardingRuleByID(id)
	if err != nil {
		if !strings.Contains(err.Error(), "No match found for") {
			return makeError(err)
		}
	}

	if portForwardingRule == nil {
		return nil
	}

	params := client.CSClient.Firewall.NewDeletePortForwardingRuleParams(id)

	_, err = client.CSClient.Firewall.DeletePortForwardingRule(params)
	if err != nil {
		return makeError(err)
	}

	return nil
}
