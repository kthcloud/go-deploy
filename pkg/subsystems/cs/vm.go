package cs

import (
	"fmt"
	"go-deploy/pkg/subsystems/cs/models"
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
		return nil, makeError(err)
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

	params := client.CSClient.VirtualMachine.NewDeployVirtualMachineParams(
		public.ServiceOfferingID,
		public.TemplateID,
		public.ZoneID,
	)

	params.SetName(public.Name)
	params.SetDisplayname(public.Name)
	params.SetNetworkids([]string{})
	params.SetProjectid(public.ProjectID)
	params.SetExtraconfig(public.ExtraConfig)

	vm, err := client.CSClient.VirtualMachine.DeployVirtualMachine(params)
	if err != nil {
		return "", makeError(err)
	}

	return vm.Id, nil
}

func (client *Client) UpdateVM(public *models.VmPublic) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update cs vm %s. details: %s", public.Name, err)
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
		return fmt.Errorf("failed to delete cs vm %s. details: %s", id, err)
	}

	if id == "" {
		return fmt.Errorf("id required")
	}

	params := client.CSClient.VirtualMachine.NewDestroyVirtualMachineParams(id)

	_, err := client.CSClient.VirtualMachine.DestroyVirtualMachine(params)
	if err != nil {
		return makeError(err)
	}

	return nil
}
