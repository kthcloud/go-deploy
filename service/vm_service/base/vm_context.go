package base

import (
	configModels "go-deploy/models/config"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/pkg/config"
	"go-deploy/service/resources"
)

type VmContext struct {
	VM        *vmModel.VM
	Generator *resources.PublicGeneratorType
	Zone      *configModels.VmZone
}

func NewVmBaseContext(vmID string) (*VmContext, error) {
	vm, err := vmModel.New().GetByID(vmID)
	if err != nil {
		return nil, err
	}

	if vm == nil {
		return nil, VmDeletedErr
	}

	zone := config.Config.VM.GetZone(vm.Zone)
	if zone == nil {
		return nil, ZoneNotFoundErr
	}

	return &VmContext{
		VM:        vm,
		Generator: resources.PublicGenerator().WithVmZone(zone).WithVM(vm),
		Zone:      zone,
	}, nil
}

func NewVmBaseContextWithoutVM(zoneName string) (*VmContext, error) {
	zone := config.Config.VM.GetZone(zoneName)
	if zone == nil {
		return nil, ZoneNotFoundErr
	}

	return &VmContext{
		Zone:      zone,
		Generator: resources.PublicGenerator().WithVmZone(zone),
	}, nil
}

func (vc *VmContext) Refresh() error {
	vm, err := vmModel.New().GetByID(vc.VM.ID)
	if err != nil {
		return err
	}

	if vm == nil {
		return VmDeletedErr
	}

	vc.VM = vm
	vc.Generator.WithVM(vm)

	return nil
}
