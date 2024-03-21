package gpus

import (
	"go-deploy/models/model"
	"go-deploy/pkg/config"
	"go-deploy/pkg/db/resources/gpu_repo"
	sUtils "go-deploy/service/utils"
	"go-deploy/service/v2/vms/opts"
)

// Get gets a GPU
func (c *Client) Get(id string, opts ...opts.GetGpuOpts) (*model.GPU, error) {
	o := sUtils.GetFirstOrDefault(opts)

	var usePrivilegedGPUs bool
	if !c.V2.HasAuth() || c.V2.Auth().GetEffectiveRole().Permissions.UsePrivilegedGPUs {
		usePrivilegedGPUs = true
	}

	var excludedGpus []string
	if !usePrivilegedGPUs {
		excludedGpus = config.Config.GPU.PrivilegedGPUs
	}

	gmc := gpu_repo.New().WithExclusion(config.Config.GPU.ExcludedHosts, excludedGpus)

	if o.Zone != nil {
		gmc.WithZone(*o.Zone)
	}

	gpu, err := gmc.GetByID(id)
	if err != nil {
		return nil, err
	}

	if gpu == nil {
		return nil, nil
	}

	return gpu, nil
}

// List lists GPUs
func (c *Client) List(opts ...opts.ListGpuOpts) ([]model.GPU, error) {
	o := sUtils.GetFirstOrDefault(opts)

	excludedGPUs := config.Config.GPU.ExcludedGPUs

	if c.V2.Auth() != nil && !c.V2.Auth().GetEffectiveRole().Permissions.UsePrivilegedGPUs {
		effectiveRole := c.V2.Auth().GetEffectiveRole()

		if !effectiveRole.Permissions.UsePrivilegedGPUs {
			excludedGPUs = append(excludedGPUs, config.Config.GPU.PrivilegedGPUs...)
		}
	}

	gmc := gpu_repo.New().WithExclusion(config.Config.GPU.ExcludedHosts, excludedGPUs)

	if o.Pagination != nil {
		gmc.WithPagination(o.Pagination.Page, o.Pagination.PageSize)
	}

	if o.Zone != nil {
		gmc.WithZone(*o.Zone)
	}

	return gmc.List()
}
