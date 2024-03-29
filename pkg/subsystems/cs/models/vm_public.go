package models

import (
	"go-deploy/pkg/imp/cloudstack"
	"time"
)

type VmPublic struct {
	ID   string `bson:"id"`
	Name string `bson:"name"`

	CpuCores int `bson:"cpuCores"`
	RAM      int `bson:"ram"`

	ExtraConfig string    `bson:"extraConfig"`
	Tags        []Tag     `bson:"tags"`
	CreatedAt   time.Time `bson:"createdAt"`
}

func (vm *VmPublic) Created() bool {
	return vm.ID != ""
}

func (vm *VmPublic) IsPlaceholder() bool {
	return false
}

// CreateVmPublicFromGet converts a cloudstack.VirtualMachine to a VmPublic.
func CreateVmPublicFromGet(vm *cloudstack.VirtualMachine) *VmPublic {
	extraConfig := ""
	if value, found := vm.Details["extraconfig-1"]; found {
		extraConfig = value
	}

	tags := FromCsTags(vm.Tags)

	return &VmPublic{
		ID:          vm.Id,
		Name:        vm.Name,
		CpuCores:    vm.Cpunumber,
		RAM:         vm.Memory / 1024,
		ExtraConfig: extraConfig,
		Tags:        tags,
		CreatedAt:   formatCreatedAt(vm.Created),
	}
}

// CreateVmPublicFromCreate converts a cloudstack.DeployVirtualMachineResponse to a VmPublic.
func CreateVmPublicFromCreate(vm *cloudstack.DeployVirtualMachineResponse) *VmPublic {
	return CreateVmPublicFromGet(
		&cloudstack.VirtualMachine{
			Id:                vm.Id,
			Name:              vm.Name,
			Serviceofferingid: vm.Serviceofferingid,
			Templateid:        vm.Templateid,
			Details:           vm.Details,
			Tags:              vm.Tags,
			Created:           vm.Created,
		},
	)
}

// CreateVmPublicFromUpdate converts a cloudstack.UpdateVirtualMachineResponse to a VmPublic.
func CreateVmPublicFromUpdate(vm *cloudstack.UpdateVirtualMachineResponse) *VmPublic {
	return CreateVmPublicFromGet(
		&cloudstack.VirtualMachine{
			Id:                vm.Id,
			Name:              vm.Name,
			Serviceofferingid: vm.Serviceofferingid,
			Templateid:        vm.Templateid,
			Details:           vm.Details,
			Tags:              vm.Tags,
			Created:           vm.Created,
		},
	)
}
