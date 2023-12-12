package client

import (
	"fmt"
	gpuModel "go-deploy/models/sys/gpu"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/service"
)

type BaseClient[parent any] struct {
	p *parent

	*Context
}

func NewBaseClient[parent any](context *Context) BaseClient[parent] {
	if context == nil {
		context = &Context{
			vmStore:  make(map[string]*vmModel.VM),
			gpuStore: make(map[string]*gpuModel.GPU),
		}
	}

	return BaseClient[parent]{Context: context}
}

func (c *BaseClient[parent]) SetParent(p *parent) {
	c.p = p
}

func (c *BaseClient[parent]) SetContext(context *Context) {
	if context == nil {
		context = &Context{}
	}

	c.Context = context
}

func (c *BaseClient[parent]) VM(id string, vmc *vmModel.Client) (*vmModel.VM, error) {
	vm, ok := c.vmStore[id]
	if !ok || vm == nil {
		return c.fetchVM(id, "", vmc)
	}

	return vm, nil
}

func (c *BaseClient[parent]) VMs(vmc *vmModel.Client) ([]vmModel.VM, error) {
	// Right now we don't have a way to skip fetching when requesting a list of resources
	return c.fetchVMs(vmc)
}

func (c *BaseClient[parent]) GPU(id string, gmc *gpuModel.Client) (*gpuModel.GPU, error) {
	gpu, ok := c.gpuStore[id]
	if !ok || gpu == nil {
		return c.fetchGPU(id, gmc)
	}

	return gpu, nil
}

func (c *BaseClient[parent]) GPUs(gmc *gpuModel.Client) ([]gpuModel.GPU, error) {
	// Right now we don't have a way to skip fetching when requesting a list of resources
	return c.fetchGPUs(gmc)
}

func (c *BaseClient[parent]) WithAuth(auth *service.AuthInfo) *parent {
	c.Auth = auth
	return c.p
}

func (c *BaseClient[parent]) Refresh(id string) (*vmModel.VM, error) {
	return c.fetchVM(id, "", nil)
}

func (c *BaseClient[parent]) fetchVM(id, name string, vmc *vmModel.Client) (*vmModel.VM, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to fetch vm in service client: %w", err)
	}

	if vmc == nil {
		vmc = vmModel.New()
	}

	var vm *vmModel.VM
	if id != "" {
		var err error
		vm, err = vmc.GetByID(id)
		if err != nil {
			return nil, makeError(err)
		}
	} else if name != "" {
		var err error
		vm, err = vmc.GetByName(name)
		if err != nil {
			return nil, makeError(err)
		}
	}

	if vm == nil {
		return nil, nil
	}

	c.storeVM(vm)
	return vm, nil
}

func (c *BaseClient[parent]) fetchVMs(vmc *vmModel.Client) ([]vmModel.VM, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to fetch gpus in service client: %w", err)
	}

	if vmc == nil {
		vmc = vmModel.New()
	}

	vms, err := vmc.List()
	if err != nil {
		return nil, makeError(err)
	}

	for _, vm := range vms {
		c.storeVM(&vm)
	}

	return vms, nil
}

func (c *BaseClient[parent]) fetchGPU(id string, gmc *gpuModel.Client) (*gpuModel.GPU, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to fetch gpu in service client: %w", err)
	}

	if gmc == nil {
		gmc = gpuModel.New()
	}

	gpu, err := gmc.GetByID(id)
	if err != nil {
		return nil, makeError(err)
	}

	if gpu == nil {
		return nil, nil
	}

	c.storeGPU(gpu)
	return gpu, nil
}

func (c *BaseClient[parent]) fetchGPUs(gmc *gpuModel.Client) ([]gpuModel.GPU, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to fetch gpus in service client: %w", err)
	}

	if gmc == nil {
		gmc = gpuModel.New()
	}

	gpus, err := gmc.List()
	if err != nil {
		return nil, makeError(err)
	}

	for _, gpu := range gpus {
		c.storeGPU(&gpu)
	}

	return gpus, nil
}

func (c *BaseClient[parent]) storeVM(vm *vmModel.VM) {
	c.vmStore[vm.ID] = vm
}

func (c *BaseClient[parent]) storeGPU(gpu *gpuModel.GPU) {
	c.gpuStore[gpu.ID] = gpu
}
