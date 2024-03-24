package gpu_leases

import (
	"errors"
	"fmt"
	"go-deploy/dto/v2/body"
	"go-deploy/models/model"
	"go-deploy/pkg/db/resources/gpu_lease_repo"
	sErrors "go-deploy/service/errors"
	sUtils "go-deploy/service/utils"
	"go-deploy/service/v2/vms/opts"
)

// Get gets a GPU lease by ID
func (c *Client) Get(id string, opts ...opts.GetGpuLeaseOpts) (*model.GpuLease, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get gpu lease. details: %w", err)
	}

	glc := gpu_lease_repo.New()

	lease, err := glc.GetByID(id)
	if err != nil {
		return nil, makeError(err)
	}

	if lease == nil {
		return nil, nil
	}

	if c.V2.HasAuth() {
		// 1. User has access through being an admin
		if c.V2.Auth().IsAdmin {
			return glc.GetByID(id)
		}

		// 2. User has access through being the lease owner
		leaseByUserID, err := glc.WithUserID(c.V2.Auth().UserID).GetByID(id)
		if err != nil {
			return nil, makeError(err)
		}

		if leaseByUserID == nil {
			return nil, nil
		}

		// 3. User has access to the parent VM through a team
		hasAccess, err := c.V1.Teams().CheckResourceAccess(c.V2.Auth().UserID, lease.VmID)
		if err != nil {
			return nil, makeError(err)
		}

		if !hasAccess {
			return nil, nil
		}

		return lease, nil
	}

	// 4. No auth info was provided, return the lease
	return lease, nil
}

// List lists GPU leases for a VM
func (c *Client) List(opts ...opts.ListGpuLeaseOpts) ([]model.GpuLease, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to list gpu leases. details: %w", err)
	}

	o := sUtils.GetFirstOrDefault(opts)

	glc := gpu_lease_repo.New()

	if o.Pagination != nil {
		glc.WithPagination(o.Pagination.Page, o.Pagination.PageSize)
	}

	if o.VmID != nil {
		// Specific VM's GPU leases are requested
		if c.V2.HasAuth() {
			hasAccess, err := c.V1.Teams().CheckResourceAccess(c.V2.Auth().UserID, *o.VmID)
			if err != nil {
				return nil, makeError(err)
			}

			if !hasAccess {
				return nil, nil
			}
		}

		glc.WithVmID(*o.VmID)
	}

	if o.UserID != nil {
		// Specific user's VMs are requested
		if c.V1.HasAuth() && c.V1.Auth().UserID != *o.UserID && !c.V1.Auth().IsAdmin {
			return nil, nil
		}

		glc.WithUserID(*o.UserID)
	} else {
		// All VMs are requested
		if c.V1.HasAuth() && !c.V1.Auth().IsAdmin {
			return nil, nil
		}

		glc.WithUserID(c.V1.Auth().UserID)
	}

	return glc.List()
}

// Create creates a GPU lease
//
// The lease is not active immediately, but will be activated by the GPU lease worker
func (c *Client) Create(leaseID, vmID, userID string, dtoGpuLeaseCreate *body.GpuLeaseCreate) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create gpu lease. details: %w", err)
	}

	if c.V2.HasAuth() && !c.V2.Auth().IsAdmin {
		// Check if the user has access to the VM
		hasAccess, err := c.V1.Teams().CheckResourceAccess(c.V2.Auth().UserID, vmID)
		if err != nil {
			return makeError(err)
		}

		if !hasAccess {
			return nil
		}

		// If not admin, the user cannot lease a GPU forever
		dtoGpuLeaseCreate.LeaseForever = false
	}

	params := model.GpuLeaseCreateParams{}.FromDTO(dtoGpuLeaseCreate)

	var leaseDuration float64
	if !c.V2.HasAuth() || (params.LeaseForever && c.V2.Auth().IsAdmin) {
		leaseDuration = 1000 * 365 * 24 // 1000 years is close enough to forever, right? :)
	}

	// Find the lease duration by the user's plan
	if c.V2.HasAuth() {
		leaseDuration = c.V2.Auth().GetEffectiveRole().Quotas.GpuLeaseDuration
	}

	if leaseDuration == 0 {
		return makeError(errors.New("lease duration could not be determined"))
	}

	err := gpu_lease_repo.New().Create(leaseID, vmID, userID, params.GpuGroupName, leaseDuration)
	if err != nil {
		if errors.Is(err, gpu_lease_repo.GpuLeaseAlreadyExistsErr) {
			return makeError(sErrors.GpuLeaseAlreadyExistsErr)
		}

		return makeError(err)
	}

	return nil
}

// Delete deletes a GPU lease
func (c *Client) Delete(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete gpu lease. details: %w", err)
	}

	lease, err := c.Get(id)
	if err != nil {
		return makeError(err)
	}

	if lease == nil {
		return nil
	}

	if c.V2.HasAuth() {
		// 1. User has access through being an admin
		if c.V2.Auth().IsAdmin {
			err = gpu_lease_repo.New().DeleteByID(id)
			if err != nil {
				return makeError(err)
			}

			return nil
		}

		// 2. User has access through being the lease owner
		if lease.UserID == c.V2.Auth().UserID {
			err = gpu_lease_repo.New().DeleteByID(id)
			if err != nil {
				return makeError(err)
			}

			return nil
		}

		// 3. User has access to the parent VM through a team
		hasAccess, err := c.V1.Teams().CheckResourceAccess(c.V2.Auth().UserID, lease.VmID)
		if err != nil {
			return makeError(err)
		}

		if hasAccess {
			err = gpu_lease_repo.New().DeleteByID(id)
			if err != nil {
				return makeError(err)
			}

			return nil
		}

		return nil
	}

	// 4. No auth info was provided, delete the lease
	return gpu_lease_repo.New().DeleteByID(id)
}
