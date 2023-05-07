package cs

import (
	"fmt"
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

func (client *Client) CreatePublicIpAddress(public *models.PublicIpAddressPublic) (string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create public ip address %s. details: %s", public.IpAddress.String(), err)
	}

	listParams := client.CsClient.Address.NewListPublicIpAddressesParams()
	listParams.SetProjectid(client.ProjectID)
	listParams.SetAssociatednetworkid(client.NetworkID)

	listIPs, err := client.CsClient.Address.ListPublicIpAddresses(listParams)
	if err != nil {
		return "", makeError(err)
	}

	var ipID string

	for _, ip := range listIPs.PublicIpAddresses {
		found := false
		for _, tag := range ip.Tags {
			if tag.Key == "deployName" && tag.Value == public.Name {
				ipID = ip.Id
				found = true
				break
			}
		}
		if found {
			break
		}
	}

	if ipID == "" {
		createIpAddressParams := client.CsClient.Address.NewAssociateIpAddressParams()
		createIpAddressParams.SetNetworkid(client.NetworkID)
		createIpAddressParams.SetProjectid(client.ProjectID)

		created, err := client.CsClient.Address.AssociateIpAddress(createIpAddressParams)
		if err != nil {
			return "", makeError(err)
		}

		ipID = created.Id
	}

	err = client.AssertPublicIPAddressTags(ipID, public.Tags)
	if err != nil {
		return "", makeError(err)
	}

	return ipID, nil
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
