package base

import (
	"go-deploy/models/sys/enviroment"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/pkg/conf"
	"go-deploy/service/resources"
)

type VmContext struct {
	VM        *vmModel.VM
	Generator *resources.PublicGeneratorType
	Zone      *enviroment.VmZone

	CreateParams *vmModel.CreateParams
	UpdateParams *vmModel.UpdateParams
}

func NewVmBaseContext(vmID string) (*VmContext, error) {
	vm, err := vmModel.New().GetByID(vmID)
	if err != nil {
		return nil, err
	}

	if vm == nil {
		return nil, VmDeletedErr
	}

	zone := conf.Env.VM.GetZone(vm.Zone)
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
	zone := conf.Env.VM.GetZone(zoneName)
	if zone == nil {
		return nil, ZoneNotFoundErr
	}

	return &VmContext{
		Zone:      zone,
		Generator: resources.PublicGenerator().WithVmZone(zone),
	}, nil
}

func (c *VmContext) WithCreateParams(params *vmModel.CreateParams) *VmContext {
	c.CreateParams = params
	return c
}

func (c *VmContext) WithUpdateParams(params *vmModel.UpdateParams) *VmContext {
	c.UpdateParams = params
	return c
}
