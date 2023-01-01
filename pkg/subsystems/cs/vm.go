package cs

import (
	"fmt"
	"github.com/apache/cloudstack-go/v2/cloudstack"
	"go-deploy/pkg/subsystems/cs/models"
	"strings"
)

func getKeyPairName(vmName string) string {
	name := fmt.Sprintf("%s-pk", vmName)
	return name
}

func (client *Client) ReadVM(id string) (*models.VmPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create cs vm %s. details: %s", id, err)
	}

	if id == "" {
		return nil, fmt.Errorf("id required")
	}

	vm, _, err := client.CsClient.VirtualMachine.GetVirtualMachineByID(id)
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

	listVmParams := client.CsClient.VirtualMachine.NewListVirtualMachinesParams()
	listVmParams.SetName(public.Name)
	listVmParams.SetProjectid(public.ProjectID)

	vm, err := client.CsClient.VirtualMachine.ListVirtualMachines(listVmParams)
	if err != nil {
		return "", makeError(err)
	}

	if vm.Count != 0 {
		return vm.VirtualMachines[0].Id, nil
	}

	createVmParams := client.CsClient.VirtualMachine.NewDeployVirtualMachineParams(
		public.ServiceOfferingID,
		public.TemplateID,
		public.ZoneID,
	)

	createVmParams.SetName(public.Name)
	createVmParams.SetDisplayname(public.Name)
	createVmParams.SetNetworkids([]string{public.NetworkID})
	createVmParams.SetProjectid(public.ProjectID)
	createVmParams.SetExtraconfig(public.ExtraConfig)

	created, err := client.CsClient.VirtualMachine.DeployVirtualMachine(createVmParams)
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

	params := client.CsClient.VirtualMachine.NewUpdateVirtualMachineParams(public.ID)

	params.SetName(public.Name)
	params.SetDisplayname(public.Name)
	params.SetExtraconfig(public.ExtraConfig)

	_, err := client.CsClient.VirtualMachine.UpdateVirtualMachine(params)
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

	vm, _, err := client.CsClient.VirtualMachine.GetVirtualMachineByID(id)
	if err != nil {
		if !strings.Contains(err.Error(), "No match found for") {
			return makeError(err)
		}
	}

	if vm == nil {
		return nil
	}

	// delete any associated ssh key pair
	err = client.deleteKeyPairForVM(vm)
	if err != nil {
		return makeError(err)
	}

	if vm.State == "Stopping" || vm.State == "DestroyRequested" || vm.State == "Expunging" {
		return nil
	}

	params := client.CsClient.VirtualMachine.NewDestroyVirtualMachineParams(id)

	params.SetExpunge(true)

	_, err = client.CsClient.VirtualMachine.DestroyVirtualMachine(params)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func (client *Client) GetVmStatus(id string) (string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get vm %s status. details: %s", id, err)
	}

	vm, _, err := client.CsClient.VirtualMachine.GetVirtualMachineByID(id)
	if err != nil {
		return "", makeError(err)
	}

	return vm.State, nil
}

func (client *Client) AddKeyPairToVM(id, publicKey string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to add ssh keypair to vm %s. details: %s", id, err)
	}

	vm, _, err := client.CsClient.VirtualMachine.GetVirtualMachineByID(id)
	if err != nil {
		return makeError(err)
	}

	keyPairName := getKeyPairName(vm.Name)

	err = client.deleteKeyPairForVM(vm)
	if err != nil {
		return makeError(err)
	}

	registerKeyPairParams := client.CsClient.SSH.NewRegisterSSHKeyPairParams(keyPairName, publicKey)
	registerKeyPairParams.SetProjectid(vm.Projectid)
	registeredKeyPair, err := client.CsClient.SSH.RegisterSSHKeyPair(registerKeyPairParams)
	if err != nil {
		return makeError(err)
	}

	keyPair := &cloudstack.SSHKeyPair{
		Id:   registeredKeyPair.Id,
		Name: registeredKeyPair.Name,
	}

	resetKeyPairParams := client.CsClient.SSH.NewResetSSHKeyForVirtualMachineParams(id, keyPair.Name)
	resetKeyPairParams.SetProjectid(vm.Projectid)

	_, err = client.CsClient.SSH.ResetSSHKeyForVirtualMachine(resetKeyPairParams)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func (client *Client) deleteKeyPairForVM(vm *cloudstack.VirtualMachine) error {
	keyPairName := getKeyPairName(vm.Name)

	listKeyPairParams := client.CsClient.SSH.NewListSSHKeyPairsParams()
	listKeyPairParams.SetProjectid(vm.Projectid)
	listKeyPairParams.SetName(keyPairName)

	keyPairsResponse, err := client.CsClient.SSH.ListSSHKeyPairs(listKeyPairParams)
	if err != nil {
		return err
	}

	for _, key := range keyPairsResponse.SSHKeyPairs {
		deleteKeyPairParams := client.CsClient.SSH.NewDeleteSSHKeyPairParams(key.Name)
		deleteKeyPairParams.SetProjectid(vm.Projectid)
		_, err = client.CsClient.SSH.DeleteSSHKeyPair(deleteKeyPairParams)
		if err != nil {
			return err
		}
	}

	return nil
}
