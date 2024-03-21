package gpu_leases

import (
	"errors"
	"go-deploy/models/model"
	"go-deploy/pkg/db/resources/gpu_lease_repo"
	sUtils "go-deploy/service/utils"
	"go-deploy/service/v2/vms/opts"
)

// List lists GPU leases
func (c *Client) List(opts ...opts.ListGpuLeaseOpts) ([]model.GpuLease, error) {
	o := sUtils.GetFirstOrDefault(opts)

	glc := gpu_lease_repo.New()

	if o.Pagination != nil {
		glc.WithPagination(o.Pagination.Page, o.Pagination.PageSize)
	}

	return glc.List()
}

// Create creates a GPU lease
//
// The lease is not active immediately, but will be activated by the GPU lease worker
func (c *Client) Create(leaseID, vmID string, userID, gpuGroupName string, opts ...opts.CreateGpuLeaseOpts) error {
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

	err := gpu_lease_repo.New().Create(leaseID, vmID, userID, gpuGroupName, leaseDuration)
	if err != nil {
		if errors.Is(err, gpu_lease_repo.GpuLeaseAlreadyExistsErr) {

			return err
		}

	}

	return nil
}
