package cs

import (
	"fmt"
	"go-deploy/pkg/imp/cloudstack"
	"go-deploy/pkg/subsystems/cs/models"
	"go-deploy/utils/subsystemutils"
	"strings"
)

func (client *Client) ReadServiceOffering(id string) (*models.ServiceOfferingPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read service offering %s. details: %s", id, err)
	}

	if id == "" {
		return nil, fmt.Errorf("id required")
	}

	serviceOffering, _, err := client.CsClient.ServiceOffering.GetServiceOfferingByID(id)
	if err != nil {
		errString := err.Error()
		if !strings.Contains(errString, "No match found for") {
			return nil, makeError(err)
		}
	}

	if serviceOffering == nil {
		return nil, nil
	}

	return models.CreateServiceOfferingPublicFromGet(serviceOffering), nil
}

func (client *Client) CreateServiceOffering(public *models.ServiceOfferingPublic) (string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create service offering. details: %s", err)
	}

	params := client.CsClient.ServiceOffering.NewListServiceOfferingsParams()
	params.SetName(public.Name)
	params.SetListall(true)

	serviceOfferings, err := client.CsClient.ServiceOffering.ListServiceOfferings(params)
	if err != nil {
		return "", makeError(err)
	}

	if len(serviceOfferings.ServiceOfferings) > 0 {
		return serviceOfferings.ServiceOfferings[0].Id, nil
	}

	createParams := cloudstack.CreateServiceOfferingParams{}
	createParams.SetName(subsystemutils.GetPrefixedName(public.Name))
	createParams.SetDisplaytext(public.Name)
	createParams.SetCpunumber(public.CpuCores)
	createParams.SetCpuspeed(2048)
	createParams.SetMemory(public.RAM * 1024)
	createParams.SetOfferha(false)
	createParams.SetLimitcpuuse(false)
	createParams.SetRootdisksize(int64(public.DiskSize))

	serviceOffering, err := client.CsClient.ServiceOffering.CreateServiceOffering(&createParams)
	if err != nil {
		return "", makeError(err)
	}

	return serviceOffering.Id, nil
}

func (client *Client) DeleteServiceOffering(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete service offering %s. details: %s", id, err)
	}

	if id == "" {
		return nil
	}

	params := client.CsClient.ServiceOffering.NewDeleteServiceOfferingParams(id)

	_, err := client.CsClient.ServiceOffering.DeleteServiceOffering(params)
	if err != nil {
		return makeError(err)
	}

	return nil
}
