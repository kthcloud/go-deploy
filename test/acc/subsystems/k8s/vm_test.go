package k8s

import (
	"github.com/kthcloud/go-deploy/test"
	"github.com/kthcloud/go-deploy/test/acc"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCreateVM(t *testing.T) {
	if !acc.VmTestsEnabled {
		t.Skip("vm tests are disabled")
	}

	t.Parallel()

	c, _ := withContext(t)
	withDefaultVM(t, c)
}

func TestUpdateVM(t *testing.T) {
	if !acc.VmTestsEnabled {
		t.Skip("vm tests are disabled")
	}

	t.Parallel()

	c, _ := withContext(t)
	vm := withDefaultVM(t, c)

	vm.CpuCores = 2
	vm.RAM = 8

	vmUpdated, err := c.UpdateVM(vm)
	test.NoError(t, err, "failed to update vm")

	assert.Equal(t, vm.CpuCores, vmUpdated.CpuCores, "vm cpu cores does not match")
	assert.Equal(t, vm.RAM, vmUpdated.RAM, "vm ram does not match")
}
