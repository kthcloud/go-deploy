package vms

import (
	gpuModels "go-deploy/models/sys/gpu"
	"go-deploy/pkg/config"
	sUtils "go-deploy/service/utils"
	"go-deploy/service/v2/vms/opts"
)

// GetGPU gets an existing deployment.
//
// It can be fetched in multiple ways including ID, name, transfer code, and Harbor webhook.
// It supports service.AuthInfo, and will restrict the result to ensure the user has access to the resource.
func (c *Client) GetGPU(id string, opts ...opts.GetGpuOpts) (*gpuModels.GPU, error) {
	o := sUtils.GetFirstOrDefault(opts)

	var usePrivilegedGPUs bool
	if !c.V2.HasAuth() || c.V2.Auth().GetEffectiveRole().Permissions.UsePrivilegedGPUs {
		usePrivilegedGPUs = true
	}

	var excludedGpus []string
	if !usePrivilegedGPUs {
		excludedGpus = config.Config.GPU.PrivilegedGPUs
	}

	gmc := gpuModels.New().WithExclusion(config.Config.GPU.ExcludedHosts, excludedGpus)

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

// ListGPUs lists GPUs
//
// It uses service.AuthInfo to only return the resources the requesting user has access to
func (c *Client) ListGPUs(opts ...opts.ListGpuOpts) ([]gpuModels.GPU, error) {
	o := sUtils.GetFirstOrDefault(opts)

	excludedGPUs := config.Config.GPU.ExcludedGPUs

	if c.V2.Auth() != nil && !c.V2.Auth().GetEffectiveRole().Permissions.UsePrivilegedGPUs {
		effectiveRole := c.V2.Auth().GetEffectiveRole()

		if !effectiveRole.Permissions.UsePrivilegedGPUs {
			excludedGPUs = append(excludedGPUs, config.Config.GPU.PrivilegedGPUs...)
		}
	}

	gmc := gpuModels.New().WithExclusion(config.Config.GPU.ExcludedHosts, excludedGPUs)

	if o.Pagination != nil {
		gmc.WithPagination(o.Pagination.Page, o.Pagination.PageSize)
	}

	if o.Zone != nil {
		gmc.WithZone(*o.Zone)
	}

	return gmc.List()
}
