package k8s

import (
	"github.com/stretchr/testify/assert"
	"go-deploy/test"
	"testing"
)

const kubevirtZone = "se-flem-2"

func TestCreateVM(t *testing.T) {
	t.Parallel()

	c, _ := withContext(t, kubevirtZone)
	withDefaultVM(t, c)
}

func TestUpdateVM(t *testing.T) {
	t.Parallel()

	c, _ := withContext(t, kubevirtZone)
	vm := withDefaultVM(t, c)

	vm.CpuCores = 2
	vm.RAM = 8

	vmUpdated, err := c.UpdateVM(vm)
	test.NoError(t, err, "failed to update vm")

	assert.Equal(t, vm.CpuCores, vmUpdated.CpuCores, "vm cpu cores does not match")
	assert.Equal(t, vm.RAM, vmUpdated.RAM, "vm ram does not match")
}
