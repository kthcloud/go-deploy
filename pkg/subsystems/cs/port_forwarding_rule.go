package cs

import (
	"errors"
	"fmt"
	"go-deploy/pkg/subsystems/cs/models"
	"strconv"
	"strings"
)

func (client *Client) ReadPortForwardingRules(vmID string) ([]models.PortForwardingRulePublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read port forwarding rules for vm %s. details: %s", vmID, err)
	}

	if vmID == "" {
		return nil, makeError(errors.New("vm id required"))
	}

	vm, _, err := client.CsClient.VirtualMachine.GetVirtualMachineByID(vmID)
	if err != nil {
		return nil, makeError(err)
	}

	listRulesParams := client.CsClient.Firewall.NewListPortForwardingRulesParams()
	listRulesParams.SetProjectid(vm.Projectid)
	listRulesParams.SetListall(true)

	portForwardingRulesResponse, err := client.CsClient.Firewall.ListPortForwardingRules(listRulesParams)
	if err != nil {
		return nil, makeError(err)
	}

	var publicRules []models.PortForwardingRulePublic
	for _, rule := range portForwardingRulesResponse.PortForwardingRules {
		if rule.Virtualmachineid != vmID {
			continue
		}

		publicRules = append(publicRules, *models.CreatePortForwardingRulePublicFromGet(rule))
	}

	return publicRules, nil
}

func (client *Client) ReadPortForwardingRule(id string) (*models.PortForwardingRulePublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read port forwarding rule %s. details: %s", id, err)
	}

	if id == "" {
		return nil, fmt.Errorf("id required")
	}

	portForwardingRule, _, err := client.CsClient.Firewall.GetPortForwardingRuleByID(id)
	if err != nil {
		if !strings.Contains(err.Error(), "No match found for") {
			return nil, makeError(err)
		}
	}

	var public *models.PortForwardingRulePublic
	if portForwardingRule != nil {
		public = models.CreatePortForwardingRulePublicFromGet(portForwardingRule)
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

	if public.IpAddressID == "" {
		return "", makeError(errors.New("ip address id required"))
	}

	listRulesParams := client.CsClient.Firewall.NewListPortForwardingRulesParams()
	listRulesParams.SetProjectid(client.ProjectID)
	listRulesParams.SetNetworkid(client.NetworkID)
	listRulesParams.SetIpaddressid(public.IpAddressID)
	listRulesParams.SetListall(true)

	listRules, err := client.CsClient.Firewall.ListPortForwardingRules(listRulesParams)
	if err != nil {
		return "", makeError(err)
	}

	publicPortStr := strconv.Itoa(public.PublicPort)
	privatePortStr := strconv.Itoa(public.PrivatePort)

	var ruleID string

	for _, rule := range listRules.PortForwardingRules {
		if rule.Virtualmachineid == public.VmID &&
			rule.Publicport == publicPortStr && rule.Privateport == privatePortStr &&
			strings.ToLower(rule.Protocol) == strings.ToLower(public.Protocol) {
			ruleID = rule.Id
		}
	}

	if ruleID == "" {
		createRuleParams := client.CsClient.Firewall.NewCreatePortForwardingRuleParams(
			public.IpAddressID,
			public.PrivatePort,
			public.Protocol,
			public.PublicPort,
			public.VmID,
		)

		created, err := client.CsClient.Firewall.CreatePortForwardingRule(createRuleParams)
		if err != nil {
			return "", makeError(err)
		}

		ruleID = created.Id
	}

	err = client.AssertPortForwardingRulesTags(ruleID, public.Tags)
	if err != nil {
		return "", makeError(err)
	}

	return ruleID, nil
}

func (client *Client) DeletePortForwardingRule(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete port forwarding rule %s. details: %s", id, err)
	}

	if id == "" {
		return fmt.Errorf("id required")
	}

	portForwardingRule, _, err := client.CsClient.Firewall.GetPortForwardingRuleByID(id)
	if err != nil {
		if !strings.Contains(err.Error(), "No match found for") {
			return makeError(err)
		}
	}

	if portForwardingRule == nil {
		return nil
	}

	params := client.CsClient.Firewall.NewDeletePortForwardingRuleParams(id)

	_, err = client.CsClient.Firewall.DeletePortForwardingRule(params)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func getTag(rule *models.PortForwardingRulePublic, tag string) string {
	var name string
	for _, tag := range rule.Tags {
		if tag.Key == "Name" {
			name = tag.Value
			break
		}
	}
	return name
}
