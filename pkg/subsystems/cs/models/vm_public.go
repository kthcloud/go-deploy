package models

import (
	"go-deploy/pkg/imp/cloudstack"
)

type VmPublic struct {
	ID   string `bson:"id"`
	Name string `bson:"name"`

	ServiceOfferingID string `bson:"serviceOfferingId"`
	TemplateID        string `bson:"templateId"`
	ExtraConfig       string `bson:"extraConfig"`
	Tags              []Tag  `bson:"tags"`
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
	}
}
