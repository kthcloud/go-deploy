package vms

import (
	"errors"
	gpuLeaseModels "go-deploy/models/sys/gpu_lease"
	sUtils "go-deploy/service/utils"
	"go-deploy/service/v2/vms/opts"
)

// ListGpuLeases lists GPU leases
func (c *Client) ListGpuLeases(opts ...opts.ListGpuLeaseOpts) ([]gpuLeaseModels.GpuLease, error) {
	o := sUtils.GetFirstOrDefault(opts)

	glc := gpuLeaseModels.New()

	if o.Pagination != nil {
		glc.WithPagination(o.Pagination.Page, o.Pagination.PageSize)
	}

	return glc.List()
}

// CreateGpuLease creates a GPU lease
//
// The lease is not active immediately, but will be activated by the GPU lease worker
func (c *Client) CreateGpuLease(leaseID, vmID string, userID, gpuGroupName string, opts ...opts.CreateGpuLeaseOpts) error {
	o := sUtils.GetFirstOrDefault(opts)

	var leaseDuration float64
	if !c.V2.HasAuth() || (o.LeaseForever && c.V2.Auth().IsAdmin) {
		leaseDuration = 1000 * 365 * 24 // 1000 years is close enough to forever :)
	}

	// Find the lease duration by the user's plan
	if c.V2.HasAuth() {
		leaseDuration = c.V2.Auth().GetEffectiveRole().Quotas.GpuLeaseDuration
	}

	if leaseDuration == 0 {
		return errors.New("lease duration could not be determined")
	}

	err := gpuLeaseModels.New().Create(leaseID, vmID, userID, gpuGroupName, leaseDuration)
	if err != nil {
		if errors.Is(err, gpuLeaseModels.GpuLeaseAlreadyExistsErr) {

			return err
		}

	}

	return nil
}
