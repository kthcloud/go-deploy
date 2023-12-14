package service

import (
	deploymentModels "go-deploy/models/sys/deployment"
	gpuModels "go-deploy/models/sys/gpu"
	smModels "go-deploy/models/sys/sm"
	vmModels "go-deploy/models/sys/vm"
)

type Cache struct {
	deploymentStore map[string]*deploymentModels.Deployment
	vmStore         map[string]*vmModels.VM
	gpuStore        map[string]*gpuModels.GPU
	smStore         map[string]*smModels.SM

	// AuthInfo is the authentication information for the client.
	Auth *AuthInfo
}

func NewCache() *Cache {
	return &Cache{
		deploymentStore: make(map[string]*deploymentModels.Deployment),
		vmStore:         make(map[string]*vmModels.VM),
		gpuStore:        make(map[string]*gpuModels.GPU),
		smStore:         make(map[string]*smModels.SM),
	}
}

func (c *Cache) StoreDeployment(deployment *deploymentModels.Deployment) {
	c.deploymentStore[deployment.ID] = deployment
}

func (c *Cache) GetDeployment(id string) *deploymentModels.Deployment {
	r, ok := c.deploymentStore[id]
	if !ok {
		return nil
	}

	return r
}

func (c *Cache) StoreVM(vm *vmModels.VM) {
	c.vmStore[vm.ID] = vm
}

func (c *Cache) GetVM(id string) *vmModels.VM {
	r, ok := c.vmStore[id]
	if !ok {
		return nil
	}

	return r
}

func (c *Cache) StoreGPU(gpu *gpuModels.GPU) {
	c.gpuStore[gpu.ID] = gpu
}

func (c *Cache) GetGPU(id string) *gpuModels.GPU {
	r, ok := c.gpuStore[id]
	if !ok {
		return nil
	}

	return r
}

func (c *Cache) StoreSM(sm *smModels.SM) {
	c.smStore[sm.ID] = sm
	c.smStore[sm.OwnerID] = sm
}

func (c *Cache) GetSM(id string) *smModels.SM {
	r, ok := c.smStore[id]
	if !ok {
		return nil
	}

	return r
}
