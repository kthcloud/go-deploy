package cs

import (
	"fmt"
	"go-deploy/pkg/subsystems/cs/models"
	"log"
	"sort"
	"strconv"
	"strings"
)

func (client *Client) ReadPortForwardingRule(id string) (*models.PortForwardingRulePublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read port forwarding rule %s. details: %w", id, err)
	}

	if id == "" {
		log.Println("cs port forwarding rule not supplied when reading. assuming it was deleted")
		return nil, nil
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
		return fmt.Errorf("failed to create port forwarding rule for vm %s. details: %w", public.VmID, err)
	}

	if public.VmID == "" {
		log.Println("cs vm id not supplied when creating port forwarding rule. assuming it was deleted")
		return "", nil
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

	var nameTag string
	for _, tag := range public.Tags {
		if tag.Key == "name" {
			nameTag = tag.Value
			break
		}
	}

	var ruleID string
	for _, rule := range listRules.PortForwardingRules {
		if rule.Virtualmachineid == public.VmID &&
			rule.Tags != nil {

			foundMatch := false
			for _, tag := range rule.Tags {
				if tag.Key == "name" && tag.Value == nameTag {
					ruleID = rule.Id
					foundMatch = true
					break
				}
			}

			if foundMatch {
				break
			}
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
		return fmt.Errorf("failed to delete port forwarding rule %s. details: %w", id, err)
	}

	if id == "" {
		log.Println("cs port forwarding rule", id, "not supplied when deleting. assuming it was deleted")
		return nil
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
		return fmt.Errorf("failed to get free port. details: %w", err)
	}

	listRulesParams := client.CsClient.Firewall.NewListPortForwardingRulesParams()
	listRulesParams.SetProjectid(client.ProjectID)
	listRulesParams.SetNetworkid(client.NetworkID)
	listRulesParams.SetIpaddressid(client.IpAddressID)
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
