package cs

import (
	"fmt"
	"go-deploy/pkg/subsystems/cs/models"
	"strings"
)

func (client *Client) ReadVM(id string) (*models.VmPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create cs vm %s. details: %s", id, err)
	}

	if id == "" {
		return nil, fmt.Errorf("id required")
	}

	vm, _, err := client.CSClient.VirtualMachine.GetVirtualMachineByID(id)
	if err != nil {
		if !strings.Contains(err.Error(), "No match found for") {
			return nil, makeError(err)
		}
	}

	var public *models.VmPublic
	if vm != nil {
		public = models.CreateVmPublicFromGet(vm)
	}

	return public, nil
}

func (client *Client) CreateVM(public *models.VmPublic) (string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create cs vm %s. details: %s", public.Name, err)
	}

	listVmParams := client.CSClient.VirtualMachine.NewListVirtualMachinesParams()
	listVmParams.SetName(public.Name)
	listVmParams.SetProjectid(public.ProjectID)

	vm, err := client.CSClient.VirtualMachine.ListVirtualMachines(listVmParams)
	if err != nil {
		return "", makeError(err)
	}

	if vm.Count != 0 {
		return vm.VirtualMachines[0].Id, nil
	}

	createVmParams := client.CSClient.VirtualMachine.NewDeployVirtualMachineParams(
		public.ServiceOfferingID,
		public.TemplateID,
		public.ZoneID,
	)

	createVmParams.SetName(public.Name)
	createVmParams.SetDisplayname(public.Name)
	createVmParams.SetNetworkids([]string{public.NetworkID})
	createVmParams.SetProjectid(public.ProjectID)
	createVmParams.SetExtraconfig(public.ExtraConfig)

	created, err := client.CSClient.VirtualMachine.DeployVirtualMachine(createVmParams)
	if err != nil {
		return "", makeError(err)
	}

	return created.Id, nil
}

func (client *Client) UpdateVM(public *models.VmPublic) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update vm %s. details: %s", public.Name, err)
	}

	if public.ID == "" {
		return fmt.Errorf("id required")
	}

	params := client.CSClient.VirtualMachine.NewUpdateVirtualMachineParams(public.ID)

	params.SetName(public.Name)
	params.SetDisplayname(public.Name)
	params.SetExtraconfig(public.ExtraConfig)

	_, err := client.CSClient.VirtualMachine.UpdateVirtualMachine(params)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func (client *Client) DeleteVM(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete vm %s. details: %s", id, err)
	}

	if id == "" {
		return fmt.Errorf("id required")
	}

	vm, _, err := client.CSClient.VirtualMachine.GetVirtualMachineByID(id)
	if err != nil {
		if !strings.Contains(err.Error(), "No match found for") {
			return makeError(err)
		}
	}

	if vm == nil {
		return nil
	}

	if vm.State == "Stopping" || vm.State == "DestroyRequested" || vm.State == "Expunging" {
		return nil
	}

	params := client.CSClient.VirtualMachine.NewDestroyVirtualMachineParams(id)

	params.SetExpunge(true)

	_, err = client.CSClient.VirtualMachine.DestroyVirtualMachine(params)
	if err != nil {
		return makeError(err)
	}

	return nil
}
