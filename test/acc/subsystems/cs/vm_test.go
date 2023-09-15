package cs

import (
	"github.com/stretchr/testify/assert"
	"sort"
	"testing"
)

func TestCreateVM(t *testing.T) {
	vm := withVM(t, withCsServiceOfferingType1(t))
	defer cleanUpVM(t, vm.ID)
}

func TestUpdateVM(t *testing.T) {
	client := withCsClient(t)

	vm := withVM(t, withCsServiceOfferingType1(t))
	soNew := withServiceOfferingType2(t)
	oldServiceOfferingID := vm.ServiceOfferingID

	defer func() {
		cleanUpVM(t, vm.ID)
		cleanUpServiceOffering(t, oldServiceOfferingID)
		cleanUpServiceOffering(t, soNew.ID)
	}()

	vm.ServiceOfferingID = soNew.ID
	vm.ExtraConfig = "some gpu config"

	err := client.UpdateVM(vm)
	assert.Error(t, err, "failed to update vm")

	err = client.DoVmCommand(vm.ID, nil, "stop")
	assert.NoError(t, err, "failed to stop vm")

	err = client.UpdateVM(vm)
	assert.NoError(t, err, "failed to update vm")

	updated, err := client.ReadVM(vm.ID)

	// sort tags
	sort.Slice(vm.Tags, func(i, j int) bool {
		return vm.Tags[i].Key < vm.Tags[j].Key
	})
	sort.Slice(updated.Tags, func(i, j int) bool {
		return updated.Tags[i].Key < updated.Tags[j].Key
	})

	assert.NoError(t, err, "failed to read vm after update")
	assert.NotNil(t, updated, "vm is nil after update")
	assert.EqualValues(t, vm, updated, "vm is not updated")
}
