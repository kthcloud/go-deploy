package cs

import (
	"errors"
	"fmt"
	"go-deploy/pkg/subsystems/cs/models"
	"sort"
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

	listRulesParams := client.CsClient.Firewall.NewListPortForwardingRulesParams()
	listRulesParams.SetProjectid(client.ProjectID)
	listRulesParams.SetNetworkid(client.NetworkID)
	listRulesParams.SetIpaddressid(client.IpAddressID)
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
			client.IpAddressID,
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

func (client *Client) GetFreePort(startPort, endPort int) (int, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get free port. details: %s", err)
	}

	listRulesParams := client.CsClient.Firewall.NewListPortForwardingRulesParams()
	listRulesParams.SetProjectid(client.ProjectID)
	listRulesParams.SetNetworkid(client.NetworkID)
	listRulesParams.SetListall(true)

	listRules, err := client.CsClient.Firewall.ListPortForwardingRules(listRulesParams)
	if err != nil {
		return -1, nil
	}

	var ports []int
	for _, rule := range listRules.PortForwardingRules {
		port, err := strconv.Atoi(rule.Publicport)
		if err != nil {
			return 0, makeError(err)
		}

		ports = append(ports, port)
	}

	sort.Ints(ports)

	var freePort int
	for i := startPort; i < len(ports); i++ {
		if ports[i]-ports[i-1] > 1 {
			freePort = ports[i-1] + 1
			break
		}
	}

	if len(ports) == 0 {
		freePort = startPort
	} else if freePort == 0 {
		freePort = ports[len(ports)-1] + 1
	}

	if freePort > endPort {
		return -1, nil
	}

	return freePort, nil
}
