package models

import "github.com/apache/cloudstack-go/v2/cloudstack"

type VmPublic struct {
	ID                string `bson:"id"`
	Name              string `bson:"name"`
	ServiceOfferingID string `bson:"serviceOfferingId"`
	TemplateID        string `bson:"templateId"`
	NetworkID         string `bson:"networkId"`
	ZoneID            string `bson:"zoneId"`
	ProjectID         string `bson:"projectId"`
	ExtraConfig       string `bson:"extraConfig"`
}

func CreateVmPublicFromGet(vm *cloudstack.VirtualMachine) *VmPublic {
	extraConfig := ""
	if value, found := vm.Details["extraconfig-1"]; found {
		extraConfig = value
	}

	return &VmPublic{
		ID:                vm.Id,
		Name:              vm.Name,
		ServiceOfferingID: vm.Serviceofferingid,
		TemplateID:        vm.Templateid,
		NetworkID:         vm.Nic[0].Networkid,
		ZoneID:            vm.Zoneid,
		ProjectID:         vm.Projectid,
		ExtraConfig:       extraConfig,
	}
}
