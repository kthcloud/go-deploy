package pfsense

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"go-deploy/pkg/subsystems/pfsense/models"
	"go-deploy/utils/requestutils"
	"go-deploy/utils/subsystemutils"
	"strconv"
)

func (client *Client) PortForwardingRuleCreated(name string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if pfsense port forwarding rule is created. details: %s", err.Error())
	}

	rule, err := client.ReadPortForwardingRule(name)
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

func (client *Client) fetchPortForwardingRules() ([]models.PortForwardRuleRead, error) {
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

	var rules []models.PortForwardRuleRead
	err = requestutils.ParseJson(pfSenseResponse.Data, &rules)
	if err != nil {
		return nil, makeError(err)
	}

	return rules, nil
}

func (client *Client) ReadPortForwardingRules() ([]models.PortForwardingRulePublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to list pfsense port forwarding rules. details: %s", err.Error())
	}

	rules, err := client.fetchPortForwardingRules()
	if err != nil {
		return nil, makeError(err)
	}

	var publics []models.PortForwardingRulePublic
	for _, rule := range rules {
		publics = append(publics, *models.CreatePortForwardingRulePublicFromRead(&rule))
	}

	return publics, nil
}

func (client *Client) ReadPortForwardingRule(id string) (*models.PortForwardingRulePublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get pfsense port forwarding rule. details: %s", err.Error())
	}

	rules, err := client.fetchPortForwardingRules()
	if err != nil {
		return nil, makeError(err)
	}

	for _, rule := range rules {
		ruleID, _, err := rule.GetIdAndName()
		if err != nil {
			continue
		}

		if ruleID == id {
			return models.CreatePortForwardingRulePublicFromRead(&rule), nil
		}
	}
	return nil, nil
}

func (client *Client) CreatePortForwardingRule(public *models.PortForwardingRulePublic) (string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create pfsense port forwarding rule. details: %s", err.Error())
	}

	if public.Name == "" {
		return "", makeError(errors.New("name required"))
	}

	if public.LocalAddress.String() == "" {
		return "", makeError(errors.New("local address required"))
	}

	if public.LocalPort == 0 {
		return "", makeError(errors.New("local port required"))
	}

	allRules, err := client.fetchPortForwardingRules()
	if err != nil {
		return "", makeError(err)
	}

	var result *models.PortForwardRuleRead
	for _, rule := range allRules {
		id, name, err := rule.GetIdAndName()
		if err != nil {
			continue
		}

		if name == subsystemutils.GetPrefixedName(public.Name) {
			return id, nil
		}
	}

	if result != nil {
		return "", nil
	}

	public.ExternalPort = client.getFreePublicPort()
	public.ExternalAddress = client.publicIP
	public.ID = uuid.New().String()

	requestBody := models.CreatePortForwardRuleCreateBody(public)
	response, err := client.doJSONRequest("POST", "/firewall/nat/port_forward", requestBody)
	if err != nil {
		return "", makeError(err)
	}

	pfSenseResponse, err := ParseResponse(response)
	if err != nil {
		return "", makeError(err)
	}

	if !requestutils.IsGoodStatusCode(response.StatusCode) {
		return "", makeApiError(pfSenseResponse, makeError)
	}

	var createdRule models.PortForwardRuleRead
	err = requestutils.ParseJson(pfSenseResponse.Data, &createdRule)
	if err != nil {
		return "", makeError(err)
	}

	id, _, err := createdRule.GetIdAndName()
	if err != nil {
		return "", makeError(err)
	}

	return id, nil
}

func (client *Client) DeletePortForwardingRule(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete pfsense port forwarding rule. details: %s", err.Error())
	}

	rules, err := client.fetchPortForwardingRules()
	if err != nil {
		return makeError(err)
	}

	resultIndex := -1

	for idx, rule := range rules {
		ruleID, _, err := rule.GetIdAndName()
		if err != nil {
			continue
		}

		if ruleID == id {
			resultIndex = idx
			break
		}
	}

	if resultIndex == -1 {
		return nil
	}

	response, err := client.doQueryRequest("DELETE", "/firewall/nat/port_forward", map[string]string{
		"id":    strconv.Itoa(resultIndex),
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
