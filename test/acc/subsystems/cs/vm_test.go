package cs

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/pkg/subsystems/cs/commands"
	"testing"
)

func TestCreateVM(t *testing.T) {
	t.Parallel()

	withDefaultVM(t, withCsServiceOfferingSmall(t))
}

func TestUpdateVM(t *testing.T) {
	t.Parallel()

	client := withClient(t)
	vm := withDefaultVM(t, withCsServiceOfferingSmall(t))

	vm.Name = vm.Name + "-updated"
	vm.ExtraConfig = "some-extra-config"

	vmUpdated, err := client.UpdateVM(vm)
	assert.NoError(t, err, "failed to update vm")

	assert.Equal(t, vm.Name, vmUpdated.Name, "vm name is not updated")
	assert.Equal(t, vm.ExtraConfig, vmUpdated.ExtraConfig, "vm extra config is not updated")
}

func TestUpdateVmServiceOffering(t *testing.T) {
	t.Parallel()
	
	client := withClient(t)
	soNew := withCsServiceOfferingBig(t)
	vm := withDefaultVM(t, withCsServiceOfferingSmall(t))

	err := client.DoVmCommand(vm.ID, nil, commands.Stop)
	assert.NoError(t, err, "failed to stop vm")

	vm.ServiceOfferingID = soNew.ID

	vmUpdated, err := client.UpdateVM(vm)
	assert.NoError(t, err, "failed to update vm service offering")

	err = client.DoVmCommand(vm.ID, nil, commands.Start)
	assert.NoError(t, err, "failed to start vm")

	assert.Equal(t, soNew.ID, vmUpdated.ServiceOfferingID, "vm service offering is not updated")
}
