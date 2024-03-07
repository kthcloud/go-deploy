package core

import (
	deploymentModels "go-deploy/models/sys/deployment"
	gpuModels "go-deploy/models/sys/gpu"
	gpuLeaseModels "go-deploy/models/sys/gpu_lease"
	jobModels "go-deploy/models/sys/job"
	notificationModels "go-deploy/models/sys/notification"
	smModels "go-deploy/models/sys/sm"
	teamModels "go-deploy/models/sys/team"
	userModels "go-deploy/models/sys/user"
	vmModels "go-deploy/models/sys/vm"
)

// Cache is a cache for all resources fetched inside the service.
// This is used to avoid fetching the same resource multiple times.
type Cache struct {
	deploymentStore   map[string]*deploymentModels.Deployment
	vmStore           map[string]*vmModels.VM
	gpuStore          map[string]*gpuModels.GPU
	gpuLeaseStore     map[string]*gpuLeaseModels.GpuLease
	smStore           map[string]*smModels.SM
	userStore         map[string]*userModels.User
	teamStore         map[string]*teamModels.Team
	jobStore          map[string]*jobModels.Job
	notificationStore map[string]*notificationModels.Notification
}

// NewCache creates a new cache.
func NewCache() *Cache {
	return &Cache{
		deploymentStore:   make(map[string]*deploymentModels.Deployment),
		vmStore:           make(map[string]*vmModels.VM),
		gpuStore:          make(map[string]*gpuModels.GPU),
		gpuLeaseStore:     make(map[string]*gpuLeaseModels.GpuLease),
		smStore:           make(map[string]*smModels.SM),
		userStore:         make(map[string]*userModels.User),
		teamStore:         make(map[string]*teamModels.Team),
		jobStore:          make(map[string]*jobModels.Job),
		notificationStore: make(map[string]*notificationModels.Notification),
	}
}

// StoreDeployment stores a deployment in the cache.
// It only stores the deployment if it is not nil.
func (c *Cache) StoreDeployment(deployment *deploymentModels.Deployment) {
	if deployment != nil {
		if d, ok := c.deploymentStore[deployment.ID]; ok {
			*d = *deployment
		} else {
			c.deploymentStore[deployment.ID] = deployment
		}
	}
}

// GetDeployment gets a deployment from the cache.
// It returns nil if the deployment is not found.
func (c *Cache) GetDeployment(id string) *deploymentModels.Deployment {
	r, ok := c.deploymentStore[id]
	if !ok {
		return nil
	}

	return r
}

// StoreVM stores a VM in the cache.
// It only stores the VM if it is not nil.
func (c *Cache) StoreVM(vm *vmModels.VM) {
	if vm != nil {
		if v, ok := c.vmStore[vm.ID]; ok {
			*v = *vm
		} else {
			c.vmStore[vm.ID] = vm
		}
	}
}

// GetVM gets a VM from the cache.
// It returns nil if the VM is not in the cache.
func (c *Cache) GetVM(id string) *vmModels.VM {
	r, ok := c.vmStore[id]
	if !ok {
		return nil
	}

	return r
}

// StoreGPU stores a GPU in the cache.
// It only stores the GPU if it is not nil.
func (c *Cache) StoreGPU(gpu *gpuModels.GPU) {
	if gpu != nil {
		if g, ok := c.gpuStore[gpu.ID]; ok {
			*g = *gpu
		} else {
			c.gpuStore[gpu.ID] = gpu
		}
	}
}

// GetGPU gets a GPU from the cache.
// It returns nil if the GPU is not in the cache.
func (c *Cache) GetGPU(id string) *gpuModels.GPU {
	r, ok := c.gpuStore[id]
	if !ok {
		return nil
	}

	return r
}

// StoreGpuLease stores a GpuLease in the cache.
// It only stores the GpuLease if it is not nil.
func (c *Cache) StoreGpuLease(gpuLease *gpuLeaseModels.GpuLease) {
	if gpuLease != nil {
		if g, ok := c.gpuLeaseStore[gpuLease.ID]; ok {
			*g = *gpuLease
		} else {
			c.gpuLeaseStore[gpuLease.ID] = gpuLease
		}
	}
}

// GetGpuLease gets a GpuLease from the cache.
// It returns nil if the GpuLease is not in the cache.
func (c *Cache) GetGpuLease(id string) *gpuLeaseModels.GpuLease {
	r, ok := c.gpuLeaseStore[id]
	if !ok {
		return nil
	}

	return r
}

// StoreSM stores a SM in the cache.
// It only stores the SM if it is not nil.
//
// SMs are stored by both ID and OwnerID.
func (c *Cache) StoreSM(sm *smModels.SM) {
	if sm != nil {
		if s, ok := c.smStore[sm.ID]; ok {
			*s = *sm
		} else if s, ok = c.smStore[sm.OwnerID]; ok {
			*s = *sm
		} else {
			c.smStore[sm.ID] = sm
			c.smStore[sm.OwnerID] = sm
		}
	}
}

// GetSM gets a SM from the cache.
// It returns nil if the SM is not in the cache.
//
// The ID can be either the SM ID or the OwnerID.
func (c *Cache) GetSM(id string) *smModels.SM {
	r, ok := c.smStore[id]
	if !ok {
		return nil
	}

	return r
}

// StoreUser stores a user in the cache.
// It only stores the user if it is not nil.
func (c *Cache) StoreUser(user *userModels.User) {
	if user != nil {
		if u, ok := c.userStore[user.ID]; ok {
			*u = *user
		} else {
			c.userStore[user.ID] = user
		}
	}
}

// GetUser gets a user from the cache.
// It returns nil if the user is not in the cache.
func (c *Cache) GetUser(id string) *userModels.User {
	r, ok := c.userStore[id]
	if !ok {
		return nil
	}

	return r
}

// StoreTeam stores a team in the cache.
// It only stores the team if it is not nil.
func (c *Cache) StoreTeam(team *teamModels.Team) {
	if team != nil {
		if t, ok := c.teamStore[team.ID]; ok {
			*t = *team
		} else {
			c.teamStore[team.ID] = team
		}
	}
}

// GetTeam gets a team from the cache.
// It returns nil if the team is not in the cache.
func (c *Cache) GetTeam(id string) *teamModels.Team {
	r, ok := c.teamStore[id]
	if !ok {
		return nil
	}

	return r
}

// StoreJob stores a job in the cache.
// It only stores the job if it is not nil.
func (c *Cache) StoreJob(job *jobModels.Job) {
	if job != nil {
		if j, ok := c.jobStore[job.ID]; ok {
			*j = *job
		} else {
			c.jobStore[job.ID] = job
		}
	}
}

// GetJob gets a job from the cache.
// It returns nil if the job is not in the cache.
func (c *Cache) GetJob(id string) *jobModels.Job {
	r, ok := c.jobStore[id]
	if !ok {
		return nil
	}

	return r
}

// StoreNotification stores a notification in the cache.
// It only stores the notification if it is not nil.
func (c *Cache) StoreNotification(notification *notificationModels.Notification) {
	if notification != nil {
		if n, ok := c.notificationStore[notification.ID]; ok {
			*n = *notification
		} else {
			c.notificationStore[notification.ID] = notification
		}
	}
}

// GetNotification gets a notification from the cache.
// It returns nil if the notification is not in the cache.
func (c *Cache) GetNotification(id string) *notificationModels.Notification {
	r, ok := c.notificationStore[id]
	if !ok {
		return nil
	}

	return r
}
