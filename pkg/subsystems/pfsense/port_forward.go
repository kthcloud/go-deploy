package pfsense

import (
	"fmt"
	"go-deploy/pkg/subsystems/pfsense/models"
	"go-deploy/utils/requestutils"
	"net"
	"strconv"
)

func (client *Client) PortForwardingRuleCreated(name string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if pfsense port forwarding rule is created. details: %s", err.Error())
	}

	rule, err := client.GetPortForwardingRule(name)
	if err != nil {
		return false, makeError(err)
	}

	return rule != nil, nil
}

func (client *Client) PortForwardingRuleDeleted(name string) (bool, error) {
	created, err := client.PortForwardingRuleCreated(name)
	if err != nil {
		return false, err
	}

	return !created, nil
}

func (client *Client) GetPortForwardingRules() ([]models.PortForwardRule, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to list pfsense port forwarding rules. details: %s", err.Error())
	}

	response, err := client.doRequest("GET", "/firewall/nat/port_forward")
	if err != nil {
		return nil, makeError(err)
	}

	pfSenseResponse, err := ParseResponse(response)
	if err != nil {
		return nil, makeError(err)
	}

	if !requestutils.IsGoodStatusCode(response.StatusCode) {
		return nil, makeApiError(pfSenseResponse, makeError)
	}

	var rules []models.PortForwardRule
	err = requestutils.ParseJson(pfSenseResponse.Data, &rules)
	if err != nil {
		return nil, makeError(err)
	}

	return rules, nil
}

func (client *Client) GetPortForwardingRule(name string) (*models.PortForwardRule, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get pfsense port forwarding rule. details: %s", err.Error())
	}

	rules, err := client.GetPortForwardingRules()
	if err != nil {
		return nil, makeError(err)
	}

	for idx, rule := range rules {
		if rule.Descr == fmt.Sprintf("auto-%s", name) {
			rule.CurrentID = idx
			return &rule, nil
		}
	}
	return nil, nil
}

func (client *Client) CreatePortForwardingRule(name string, ip net.IP, internalPort int) (*models.PortForwardRule, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create pfsense port forwarding rule. details: %s", err.Error())
	}

	rule, err := client.GetPortForwardingRule(name)
	if err != nil {
		return nil, makeError(err)
	}

	if rule != nil {
		return nil, nil
	}

	externalPort := client.getFreePublicPort()

	requestBody := models.CreatePortForwardRuleCreateReq(client.publicIP, externalPort, ip, internalPort, fmt.Sprintf("auto-%s", name))
	response, err := client.doJSONRequest("POST", "/firewall/nat/port_forward", requestBody)
	if err != nil {
		return nil, makeError(err)
	}

	pfSenseResponse, err := ParseResponse(response)
	if err != nil {
		return nil, makeError(err)
	}

	if !requestutils.IsGoodStatusCode(response.StatusCode) {
		return nil, makeApiError(pfSenseResponse, makeError)
	}

	var createdRule models.PortForwardRule
	err = requestutils.ParseJson(pfSenseResponse.Data, &createdRule)
	if err != nil {
		return nil, makeError(err)
	}

	return &createdRule, nil
}

func (client *Client) DeletePortForwardingRule(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete pfsense port forwarding rule. details: %s", err.Error())
	}

	rule, err := client.GetPortForwardingRule(name)
	if err != nil {
		return makeError(err)
	}

	if rule == nil {
		return nil
	}

	response, err := client.doQueryRequest("DELETE", "/firewall/nat/port_forward", map[string]string{
		"id":    strconv.Itoa(rule.CurrentID),
		"apply": "true",
	})
	if err != nil {
		return makeError(err)
	}

	pfSenseResponse, err := ParseResponse(response)
	if err != nil {
		return makeError(err)
	}

	if !requestutils.IsGoodStatusCode(response.StatusCode) {
		return makeApiError(pfSenseResponse, makeError)
	}

	return nil
}
