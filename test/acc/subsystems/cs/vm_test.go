package cs

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/pkg/subsystems/cs/commands"
	"go-deploy/test"
	"testing"
)

func TestCreateVM(t *testing.T) {
	t.Skip("CloudStack is too unpredictable to run this test")

	t.Parallel()

	withDefaultVM(t)
}

func TestUpdateVM(t *testing.T) {
	t.Skip("CloudStack is too unpredictable to run this test")

	t.Parallel()

	client := withClient(t)
	vm := withDefaultVM(t)

	vm.Name = vm.Name + "-updated"
	vm.ExtraConfig = "some-extra-config"

	vmUpdated, err := client.UpdateVM(vm)
	test.NoError(t, err, "failed to update vm")

	assert.Equal(t, vm.Name, vmUpdated.Name, "vm name is not updated")
	assert.Equal(t, vm.ExtraConfig, vmUpdated.ExtraConfig, "vm extra config is not updated")
}

func TestUpdateVmSpecs(t *testing.T) {
	t.Skip("CloudStack is too unpredictable to run this test")

	t.Parallel()

	client := withClient(t)
	vm := withDefaultVM(t)

	err := client.DoVmCommand(vm.ID, nil, commands.Stop)
	test.NoError(t, err, "failed to stop vm")

	vm.CpuCores += 1
	vm.RAM += 1

	vmUpdated, err := client.UpdateVM(vm)
	test.NoError(t, err, "failed to update vm service offering")

	err = client.DoVmCommand(vm.ID, nil, commands.Start)
	test.NoError(t, err, "failed to start vm")

	assert.Equal(t, vm.CpuCores, vmUpdated.CpuCores, "vm cpu cores are not updated")
	assert.Equal(t, vm.RAM, vmUpdated.RAM, "vm ram is not updated")
}
