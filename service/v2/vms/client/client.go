package client

import (
	"fmt"
	"go-deploy/models/model"
	"go-deploy/models/versions"
	"go-deploy/pkg/db/resources/gpu_lease_repo"
	"go-deploy/pkg/db/resources/gpu_repo"
	"go-deploy/pkg/db/resources/vm_repo"
	"go-deploy/service/core"
)

// BaseClient is the base client for all the subsystems client for VMs and GPUs.
type BaseClient[parent any] struct {
	p *parent

	// Cache is used to cache the resources fetched inside the service.
	Cache *core.Cache
}

// NewBaseClient creates a new BaseClient.
func NewBaseClient[parent any](cache *core.Cache) BaseClient[parent] {
	if cache == nil {
		cache = core.NewCache()
	}

	return BaseClient[parent]{Cache: cache}
}

// SetParent sets the parent of the client.
// This ensures the correct parent client is returned when calling builder methods.
func (c *BaseClient[parent]) SetParent(p *parent) {
	c.p = p
}

// VM returns the VM with the given ID.
// After a successful fetch, the VM will be cached.
func (c *BaseClient[parent]) VM(id string, vmc *vm_repo.Client) (*model.VM, error) {
	vm := c.Cache.GetVM(id)
	if vm == nil {
		return c.fetchVM(id, "", vmc)
	}

	return vm, nil
}

// VMs returns a list of VMs.
// After a successful fetch, the VMs will be cached.
func (c *BaseClient[parent]) VMs(vmc *vm_repo.Client) ([]model.VM, error) {
	// Right now we don't have a way to skip fetching when requesting a list of resources
	return c.fetchVMs(vmc)
}

// GPU returns the GPU with the given ID.
// After a successful fetch, the GPU will be cached.
func (c *BaseClient[parent]) GPU(id string, gmc *gpu_repo.Client) (*model.GPU, error) {
	gpu := c.Cache.GetGPU(id)
	if gpu == nil {
		return c.fetchGPU(id, gmc)
	}

	return gpu, nil
}

// GPUs returns a list of GPUs.
// After a successful fetch, the GPUs will be cached.
func (c *BaseClient[parent]) GPUs(gmc *gpu_repo.Client) ([]model.GPU, error) {
	// Right now we don't have a way to skip fetching when requesting a list of resources
	return c.fetchGPUs(gmc)
}

// GpuLease returns the GPU lease with the given ID.
// After a successful fetch, the GPU lease will be cached.
func (c *BaseClient[parent]) GpuLease(id string, glc *gpu_lease_repo.Client) (*model.GpuLease, error) {
	gpuLease := c.Cache.GetGpuLease(id)
	if gpuLease == nil {
		return c.fetchGpuLease(id, glc)
	}

	return gpuLease, nil
}

// Refresh refreshes the VM with the given ID.
// After a successful fetch, the VM will be cached.
func (c *BaseClient[parent]) Refresh(id string) (*model.VM, error) {
	return c.fetchVM(id, "", nil)
}

// fetchVM fetches a VM by ID or name.
// After a successful fetch, the VM will be cached.
func (c *BaseClient[parent]) fetchVM(id, name string, vmc *vm_repo.Client) (*model.VM, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to fetch vm in service client: %w", err)
	}

	if vmc == nil {
		vmc = vm_repo.New(versions.V2)
	}

	var vm *model.VM
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

// fetchVMs fetches all VMs according to the given vmc.
// After a successful fetch, the VMs will be cached.
func (c *BaseClient[parent]) fetchVMs(vmc *vm_repo.Client) ([]model.VM, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to fetch gpus in service client: %w", err)
	}

	if vmc == nil {
		vmc = vm_repo.New(versions.V2)
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

// fetchGPU fetches a GPU by ID.
// After a successful fetch, the GPU will be cached.
func (c *BaseClient[parent]) fetchGPU(id string, gmc *gpu_repo.Client) (*model.GPU, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to fetch gpu_repo in service client: %w", err)
	}

	if gmc == nil {
		gmc = gpu_repo.New()
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

// fetchGPUs fetches all GPUs according to the given gmc.
// After a successful fetch, the GPUs will be cached.
func (c *BaseClient[parent]) fetchGPUs(gmc *gpu_repo.Client) ([]model.GPU, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to fetch gpus in service client: %w", err)
	}

	if gmc == nil {
		gmc = gpu_repo.New()
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

// fetchGpuLease fetches a GPU lease by ID.
// After a successful fetch, the GPU lease will be cached.
func (c *BaseClient[parent]) fetchGpuLease(id string, glc *gpu_lease_repo.Client) (*model.GpuLease, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to fetch gpu_repo lease in service client: %w", err)
	}

	if glc == nil {
		glc = gpu_lease_repo.New()
	}

	gpuLease, err := glc.GetByID(id)
	if err != nil {
		return nil, makeError(err)
	}

	if gpuLease == nil {
		return nil, nil
	}

	c.Cache.StoreGpuLease(gpuLease)
	return gpuLease, nil
}

// fetchGpuLeases fetches all GPU leases according to the given glc.
// After a successful fetch, the GPU leases will be cached.
func (c *BaseClient[parent]) fetchGpuLeases(glc *gpu_lease_repo.Client) ([]model.GpuLease, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to fetch gpu_repo leases in service client: %w", err)
	}

	if glc == nil {
		glc = gpu_lease_repo.New()
	}

	gpuLeases, err := glc.List()
	if err != nil {
		return nil, makeError(err)
	}

	for _, gpuLease := range gpuLeases {
		c.Cache.StoreGpuLease(&gpuLease)
	}

	return gpuLeases, nil
}
