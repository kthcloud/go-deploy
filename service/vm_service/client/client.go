package client

import (
	"fmt"
	gpuModels "go-deploy/models/sys/gpu"
	vmModels "go-deploy/models/sys/vm"
	"go-deploy/service"
)

type BaseClient[parent any] struct {
	p *parent

	// Cache is used to cache the resources fetched inside the service.
	Cache *service.Cache

	// Auth is the authentication information for the client.
	Auth *service.AuthInfo
}

func NewBaseClient[parent any](cache *service.Cache) BaseClient[parent] {
	if cache == nil {
		cache = service.NewCache()
	}

	return BaseClient[parent]{Cache: cache}
}

func (c *BaseClient[parent]) SetParent(p *parent) {
	c.p = p
}

func (c *BaseClient[parent]) VM(id string, vmc *vmModels.Client) (*vmModels.VM, error) {
	vm := c.Cache.GetVM(id)
	if vm == nil {
		return c.fetchVM(id, "", vmc)
	}

	return vm, nil
}

func (c *BaseClient[parent]) VMs(vmc *vmModels.Client) ([]vmModels.VM, error) {
	// Right now we don't have a way to skip fetching when requesting a list of resources
	return c.fetchVMs(vmc)
}

func (c *BaseClient[parent]) GPU(id string, gmc *gpuModels.Client) (*gpuModels.GPU, error) {
	gpu := c.Cache.GetGPU(id)
	if gpu == nil {
		return c.fetchGPU(id, gmc)
	}

	return gpu, nil
}

func (c *BaseClient[parent]) GPUs(gmc *gpuModels.Client) ([]gpuModels.GPU, error) {
	// Right now we don't have a way to skip fetching when requesting a list of resources
	return c.fetchGPUs(gmc)
}

func (c *BaseClient[parent]) WithAuth(auth *service.AuthInfo) *parent {
	c.Auth = auth
	return c.p
}

func (c *BaseClient[parent]) Refresh(id string) (*vmModels.VM, error) {
	return c.fetchVM(id, "", nil)
}

func (c *BaseClient[parent]) fetchVM(id, name string, vmc *vmModels.Client) (*vmModels.VM, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to fetch vm in service client: %w", err)
	}

	if vmc == nil {
		vmc = vmModels.New()
	}

	var vm *vmModels.VM
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

	c.Cache.StoreVM(vm)
	return vm, nil
}

func (c *BaseClient[parent]) fetchVMs(vmc *vmModels.Client) ([]vmModels.VM, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to fetch gpus in service client: %w", err)
	}

	if vmc == nil {
		vmc = vmModels.New()
	}

	vms, err := vmc.List()
	if err != nil {
		return nil, makeError(err)
	}

	for _, vm := range vms {
		v := vm
		c.Cache.StoreVM(&v)
	}

	return vms, nil
}

func (c *BaseClient[parent]) fetchGPU(id string, gmc *gpuModels.Client) (*gpuModels.GPU, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to fetch gpu in service client: %w", err)
	}

	if gmc == nil {
		gmc = gpuModels.New()
	}

	gpu, err := gmc.GetByID(id)
	if err != nil {
		return nil, makeError(err)
	}

	if gpu == nil {
		return nil, nil
	}

	c.Cache.StoreGPU(gpu)
	return gpu, nil
}

func (c *BaseClient[parent]) fetchGPUs(gmc *gpuModels.Client) ([]gpuModels.GPU, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to fetch gpus in service client: %w", err)
	}

	if gmc == nil {
		gmc = gpuModels.New()
	}

	gpus, err := gmc.List()
	if err != nil {
		return nil, makeError(err)
	}

	for _, gpu := range gpus {
		c.Cache.StoreGPU(&gpu)
	}

	return gpus, nil
}
