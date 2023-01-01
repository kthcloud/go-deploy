package cs

import (
	"fmt"
	"github.com/apache/cloudstack-go/v2/cloudstack"
	"go-deploy/pkg/subsystems/cs/models"
	"strings"
)

func (client *Client) ReadPublicIpAddress(id string) (*models.PublicIpAddressPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read public ip address %s. details: %s", id, err)
	}

	if id == "" {
		return nil, fmt.Errorf("id required")
	}

	publicIP, _, err := client.CsClient.Address.GetPublicIpAddressByID(id)
	if err != nil {
		if !strings.Contains(err.Error(), "No match found for") {
			return nil, makeError(err)
		}
	}

	var public *models.PublicIpAddressPublic
	if publicIP != nil {
		public = models.CreatePublicIpAddressPublicFromGet(publicIP)
	}

	return public, nil
}

func (client *Client) ReadFreePublicIpAddress(networkID, projectID string) (*models.PublicIpAddressPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read free public ip address. details: %s", err)
	}

	params := client.CsClient.Address.NewListPublicIpAddressesParams()
	params.SetProjectid(projectID)
	params.SetAssociatednetworkid(networkID)
	params.SetListall(true)

	response, err := client.CsClient.Address.ListPublicIpAddresses(params)
	if err != nil {
		return nil, makeError(err)
	}

	var publicIP *cloudstack.PublicIpAddress
	for _, ip := range response.PublicIpAddresses {
		if ip.Issourcenat {
			continue
		}

		listRulesParams := client.CsClient.Firewall.NewListPortForwardingRulesParams()
		listRulesParams.SetIpaddressid(ip.Id)
		listRulesParams.SetNetworkid(networkID)
		listRulesParams.SetProjectid(projectID)
		listRulesParams.SetListall(true)

		rulesResponse, err := client.CsClient.Firewall.ListPortForwardingRules(listRulesParams)
		if err != nil {
			return nil, makeError(err)
		}

		if rulesResponse.Count == 0 {
			publicIP = ip
			break
		}
	}

	var public *models.PublicIpAddressPublic
	if publicIP != nil {
		public = models.CreatePublicIpAddressPublicFromGet(publicIP)
	}

	return public, nil
}

func (client *Client) ReadPublicIpAddressByVmID(vmID, networkID, projectID string) (*models.PublicIpAddressPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read free public ip address. details: %s", err)
	}

	params := client.CsClient.Address.NewListPublicIpAddressesParams()
	params.SetProjectid(projectID)
	params.SetAssociatednetworkid(networkID)
	params.SetListall(true)

	response, err := client.CsClient.Address.ListPublicIpAddresses(params)
	if err != nil {
		return nil, makeError(err)
	}

	var publicIP *cloudstack.PublicIpAddress
	for _, ip := range response.PublicIpAddresses {
		if publicIP != nil {
			break
		}
		if ip.Issourcenat {
			continue
		}

		listRulesParams := client.CsClient.Firewall.NewListPortForwardingRulesParams()
		listRulesParams.SetIpaddressid(ip.Id)
		listRulesParams.SetNetworkid(networkID)
		listRulesParams.SetProjectid(projectID)
		listRulesParams.SetListall(true)

		rulesResponse, err := client.CsClient.Firewall.ListPortForwardingRules(listRulesParams)
		if err != nil {
			return nil, makeError(err)
		}

		for _, rule := range rulesResponse.PortForwardingRules {
			if rule.Virtualmachineid == vmID {
				publicIP = ip
				break
			}
		}
	}

	var public *models.PublicIpAddressPublic
	if publicIP != nil {
		public = models.CreatePublicIpAddressPublicFromGet(publicIP)
	}

	return public, nil
}

func (client *Client) CreatePublicIpAddress(public *models.PublicIpAddressPublic) (string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create public ip address %s. details: %s", public.IpAddress.String(), err)
	}

	if public.ProjectID == "" {
		return "", fmt.Errorf("project id required")
	}

	if public.NetworkID == "" {
		return "", fmt.Errorf("network id required")
	}

	createIpAddressParams := client.CsClient.Address.NewAssociateIpAddressParams()
	createIpAddressParams.SetNetworkid(public.NetworkID)
	createIpAddressParams.SetProjectid(public.ProjectID)

	created, err := client.CsClient.Address.AssociateIpAddress(createIpAddressParams)
	if err != nil {
		return "", makeError(err)
	}

	return created.Id, nil
}

func (client *Client) DeletePublicIpAddress(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete public ip address %s. details: %s", id, err)
	}

	if id == "" {
		return fmt.Errorf("id required")
	}

	publicIP, _, err := client.CsClient.Address.GetPublicIpAddressByID(id)
	if err != nil {
		if !strings.Contains(err.Error(), "No match found for") {
			return makeError(err)
		}
	}

	if publicIP == nil {
		return nil
	}

	params := client.CsClient.Address.NewDisassociateIpAddressParams(id)

	_, err = client.CsClient.Address.DisassociateIpAddress(params)
	if err != nil {
		return makeError(err)
	}

	return nil
}
