package cs

import (
	"fmt"
	"go-deploy/pkg/subsystems/cs/models"
)

func (client *Client) CreateVM(name string, params *models.CreateVMParams) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create cs vm %s. details: %s", name, err)
	}

	vmParams := client.CSClient.VirtualMachine.NewDeployVirtualMachineParams(
		params.ServiceOfferingID,
		params.TemplateID,
		client.ZoneID,
	)

	vmParams.SetName(name)
	vmParams.SetDisplayname(name)

	_, err := client.CSClient.VirtualMachine.DeployVirtualMachine(vmParams)
	if err != nil {
		return makeError(err)
	}

	return nil
}
