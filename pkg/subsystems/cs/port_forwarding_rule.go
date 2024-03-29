package cs

import (
	"fmt"
	"go-deploy/pkg/imp/cloudstack"
	"go-deploy/pkg/subsystems/cs/errors"
	"go-deploy/pkg/subsystems/cs/models"
	"log"
	"strconv"
	"strings"
	"time"
)

// ReadPortForwardingRule reads the port forwarding rule from CloudStack.
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

// CreatePortForwardingRule creates the port forwarding rule in CloudStack.
func (client *Client) CreatePortForwardingRule(public *models.PortForwardingRulePublic) (*models.PortForwardingRulePublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create port forwarding rule for vm %s. details: %w", public.VmID, err)
	}

	if public.VmID == "" {
		log.Println("cs vm id not supplied when creating port forwarding rule. assuming it was deleted")
		return nil, nil
	}

	listRulesParams := client.CsClient.Firewall.NewListPortForwardingRulesParams()
	listRulesParams.SetProjectid(client.ProjectID)
	listRulesParams.SetNetworkid(public.NetworkID)
	listRulesParams.SetIpaddressid(public.IpAddressID)
	listRulesParams.SetListall(true)

	listRules, err := client.CsClient.Firewall.ListPortForwardingRules(listRulesParams)
	if err != nil {
		return nil, makeError(err)
	}

	var id string
	for _, rule := range listRules.PortForwardingRules {
		intPrivatePort, err := strconv.Atoi(rule.Privateport)
		if err != nil {
			return nil, makeError(err)
		}

		if rule.Virtualmachineid == public.VmID && intPrivatePort == public.PrivatePort && rule.Protocol == public.Protocol {
			id = rule.Id
			break
		}
	}

	if id == "" {
		createRuleParams := client.CsClient.Firewall.NewCreatePortForwardingRuleParams(
			public.IpAddressID,
			public.PrivatePort,
			public.Protocol,
			public.PublicPort,
			public.VmID,
		)

		var created *cloudstack.CreatePortForwardingRuleResponse
		for i := 0; i < 5; i++ {
			// This can sometimes fail with "Undefined error" message. Retrying seems to fix it.
			created, err = client.CsClient.Firewall.CreatePortForwardingRule(createRuleParams)
			if err != nil {
				errStr := err.Error()
				if strings.Contains(errStr, "The range specified") && strings.Contains(errStr, "conflicts with rule") {
					return nil, errors.NewPortInUseError(public.PublicPort)
				}

				return nil, makeError(err)
			}
			if err == nil {
				break
			} else if !strings.Contains(err.Error(), "Undefined error") {
				return nil, makeError(err)
			}

			time.Sleep(1 * time.Second)
		}

		id = created.Id
	}

	err = client.AssertPortForwardingRulesTags(id, public.Tags)
	if err != nil {
		return nil, makeError(err)
	}

	return client.ReadPortForwardingRule(id)
}

// UpdatePortForwardingRule updates the port forwarding rule in CloudStack.
func (client *Client) UpdatePortForwardingRule(public *models.PortForwardingRulePublic) (*models.PortForwardingRulePublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create port forwarding rule for vm %s. details: %w", public.VmID, err)
	}

	if public.ID == "" {
		log.Println("cs port forwarding rule id not supplied when updating port forwarding rule. assuming it was deleted")
		return nil, nil
	}

	if public.VmID == "" {
		log.Println("cs vm id not supplied when updating port forwarding rule. assuming it was deleted")
		return nil, nil
	}

	updateRuleParams := client.CsClient.Firewall.NewUpdatePortForwardingRuleParams(public.ID)
	updateRuleParams.SetVirtualmachineid(public.VmID)
	updateRuleParams.SetPrivateport(public.PrivatePort)

	portForwardingRule, err := client.CsClient.Firewall.UpdatePortForwardingRule(updateRuleParams)
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "The range specified") && strings.Contains(errStr, "conflicts with rule") {
			return nil, errors.NewPortInUseError(public.PublicPort)
		}

		return nil, makeError(err)
	}

	err = client.AssertPortForwardingRulesTags(public.ID, public.Tags)
	if err != nil {
		return nil, makeError(err)
	}

	return models.CreatePortForwardingRulePublicFromUpdate(portForwardingRule), nil
}

// DeletePortForwardingRule deletes the port forwarding rule in CloudStack.
func (client *Client) DeletePortForwardingRule(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete port forwarding rule %s. details: %w", id, err)
	}

	if id == "" {
		log.Println("cs port forwarding rule not supplied when deleting. assuming it was deleted")
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

	var lastError error
	for i := 0; i < 5; i++ {
		// This can sometimes fail with "Unknown error" message. Retrying seems to fix it.
		_, lastError = client.CsClient.Firewall.DeletePortForwardingRule(params)
		if err == nil {
			return nil
		} else if !strings.Contains(err.Error(), "Undefined error") {
			return makeError(err)
		}

		time.Sleep(1 * time.Second)
	}

	return makeError(lastError)
}
