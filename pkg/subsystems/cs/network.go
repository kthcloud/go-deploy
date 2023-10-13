package cs

import (
	"fmt"
	"go-deploy/pkg/subsystems/cs/models"
	"log"
	"strings"
)

func (client *Client) ReadNetwork(id string) (*models.NetworkPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read network %s. details: %w", id, err)
	}

	if id == "" {
		log.Println("cs port forwarding rule not supplied when reading. assuming it was deleted")
		return nil, nil
	}

	network, _, err := client.CsClient.Network.GetNetworkByID(id)
	if err != nil {
		if !strings.Contains(err.Error(), "No match found for") {
			return nil, makeError(err)
		}
	}

	var public *models.NetworkPublic
	if network != nil {
		public = models.CreateNetworkPublicFromGet(network)
	}

	return public, nil
}

func (client *Client) GetNetworkSourceNatIpAddressID(id string) (string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read network %s. details: %w", id, err)
	}

	params := client.CsClient.Address.NewListPublicIpAddressesParams()
	params.SetProjectid(client.ProjectID)
	params.SetListall(true)
	params.SetAssociatednetworkid(id)

	ipAddressesRes, err := client.CsClient.Address.ListPublicIpAddresses(params)
	if err != nil {
		return "", makeError(err)
	}

	var ipAddressID string
	for _, ipAddress := range ipAddressesRes.PublicIpAddresses {
		if ipAddress.Issourcenat {
			ipAddressID = ipAddress.Id
		}
	}

	if ipAddressID == "" {
		return "", fmt.Errorf("no source nat ip address found for network %s", id)
	}

	return ipAddressID, nil
}

func (client *Client) CreateNetwork(public *models.NetworkPublic) (*models.NetworkPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create network %s. details: %w", public.Name, err)
	}

	if public.Name == "" {
		log.Println("cs network name not supplied when creating network. assuming it was deleted")
		return nil, nil
	}

	listNetworkParams := client.CsClient.Network.NewListNetworksParams()
	listNetworkParams.SetProjectid(client.ProjectID)
	listNetworkParams.SetListall(true)

	listNetwork, err := client.CsClient.Network.ListNetworks(listNetworkParams)
	if err != nil {
		return nil, makeError(err)
	}

	var nameTag string
	for _, tag := range public.Tags {
		if tag.Key == "name" {
			nameTag = tag.Value
		}
	}

	for _, network := range listNetwork.Networks {
		if network.Name == nameTag {
			return models.CreateNetworkPublicFromGet(network), nil
		}
	}

	createNetworkParams := client.CsClient.Network.NewCreateNetworkParams(public.Name, client.NetworkOfferingID, client.ZoneID)
	createNetworkParams.SetProjectid(client.ProjectID)
	createNetworkParams.SetDisplaytext(public.Description)
	createNetworkParams.SetNetworkofferingid(client.NetworkOfferingID)

	network, err := client.CsClient.Network.CreateNetwork(createNetworkParams)
	if err != nil {
		return nil, makeError(err)
	}

	return models.CreateNetworkPublicFromCreate(network), nil
}

func DeleteNetwork(client *Client, id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete network %s. details: %w", id, err)
	}

	if id == "" {
		log.Println("cs network id not supplied when deleting network. assuming it was deleted")
		return nil
	}

	deleteNetworkParams := client.CsClient.Network.NewDeleteNetworkParams(id)

	_, err := client.CsClient.Network.DeleteNetwork(deleteNetworkParams)
	if err != nil {
		if !strings.Contains(err.Error(), "No match found for") {
			return makeError(err)
		}
	}

	return nil
}
