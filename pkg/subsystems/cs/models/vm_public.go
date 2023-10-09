package models

import (
	"go-deploy/pkg/imp/cloudstack"
	"time"
)

type VmPublic struct {
	ID   string `bson:"id"`
	Name string `bson:"name"`

	ServiceOfferingID string    `bson:"serviceOfferingId"`
	TemplateID        string    `bson:"templateId"`
	ExtraConfig       string    `bson:"extraConfig"`
	Tags              []Tag     `bson:"tags"`
	CreatedAt         time.Time `bson:"createdAt"`
}

func (vm *VmPublic) Created() bool {
	return vm.ID != ""
}

func (vm *VmPublic) IsPlaceholder() bool {
	return false
}

func CreateVmPublicFromGet(vm *cloudstack.VirtualMachine) *VmPublic {
	extraConfig := ""
	if value, found := vm.Details["extraconfig-1"]; found {
		extraConfig = value
	}

	tags := FromCsTags(vm.Tags)

	var name string
	for _, tag := range tags {
		if tag.Key == "deployName" {
			name = tag.Value
		}
	}

	return &VmPublic{
		ID:                vm.Id,
		Name:              name,
		ServiceOfferingID: vm.Serviceofferingid,
		TemplateID:        vm.Templateid,
		ExtraConfig:       extraConfig,
		Tags:              tags,
		CreatedAt:         formatCreatedAt(vm.Created),
	}
}

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
