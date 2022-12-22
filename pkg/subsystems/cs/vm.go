package cs

import (
	"fmt"
	"go-deploy/pkg/subsystems/cs/models"
)

func (client *Client) CreateVM(public *models.VMPublic) (string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create cs vm %s. details: %s", public.Name, err)
	}

	vmParams := client.CSClient.VirtualMachine.NewDeployVirtualMachineParams(
		public.ServiceOfferingID,
		public.TemplateID,
		public.ZoneID,
	)

	vmParams.SetName(public.Name)
	vmParams.SetDisplayname(public.Name)
	vmParams.SetNetworkids([]string{})
	vmParams.SetProjectid(public.ProjectID)
	vmParams.SetExtraconfig(public.ExtraConfig)

	vm, err := client.CSClient.VirtualMachine.DeployVirtualMachine(vmParams)
	if err != nil {
		return "", makeError(err)
	}

	return vm.Id, nil
}

func (client *Client) UpdateVM(public *models.VMPublic) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create cs vm %s. details: %s", public.Name, err)
	}

	if public.ID == "" {
		return fmt.Errorf("id required")
	}

	vmParams := client.CSClient.VirtualMachine.NewUpdateVirtualMachineParams(public.ID)

	vmParams.SetName(public.Name)
	vmParams.SetDisplayname(public.Name)
	vmParams.SetExtraconfig(public.ExtraConfig)

	_, err := client.CSClient.VirtualMachine.UpdateVirtualMachine(vmParams)
	if err != nil {
		return makeError(err)
	}

	return nil
}
