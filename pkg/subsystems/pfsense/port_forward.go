package pfsense

import (
	"fmt"
	"go-deploy/pkg/subsystems/pfsense/models"
	"go-deploy/utils/requestutils"
	"net"
)

func (client *Client) GetPortForwardRules() ([]models.PortForwardRule, error) {
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

func (client *Client) CreatePortForwardRule(ip net.IP, internalPort int, description string) (int, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create pfsense port forwarding rule. details: %s", err.Error())
	}

	externalPort := client.getFreePublicPort()

	requestBody := models.CreatePortForwardRuleCreateReq(client.publicIP, externalPort, ip, internalPort, description)
	response, err := client.doJSONRequest("POST", "/firewall/nat/port_forward", requestBody)
	if err != nil {
		return -1, makeError(err)
	}

	pfSenseResponse, err := ParseResponse(response)
	if err != nil {
		return -1, makeError(err)
	}

	if !requestutils.IsGoodStatusCode(response.StatusCode) {
		return -1, makeApiError(pfSenseResponse, makeError)
	}

	return externalPort, nil
}

func (client *Client) ClosePort(port int) error {

	//params := pfsense.APIFirewallNATOutboundPortForwardDeleteParams{
	//	Id:    0,
	//	Apply: nil,
	//}
	//
	//response, err. = client.PfSenseClient.APIFirewallNATOutboundPortForwardDelete(context.TODO())

	return nil
}
