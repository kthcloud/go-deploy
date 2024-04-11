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
	"time"
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
		if lease.VmID != nil {
			hasAccess, err := c.V1.Teams().CheckResourceAccess(c.V2.Auth().UserID, *lease.VmID)
			if err != nil {
				return nil, makeError(err)
			}

			if !hasAccess {
				return nil, nil
			}
		}

		return lease, nil
	}

	// 4. No auth info was provided, return the lease
	return lease, nil
}

// GetByVmID gets a GPU lease by VM ID
func (c *Client) GetByVmID(vmID string, opts ...opts.GetGpuLeaseOpts) (*model.GpuLease, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get gpu lease. details: %w", err)
	}

	vm, err := c.V2.VMs().Get(vmID)
	if err != nil {
		return nil, makeError(err)
	}

	if vm == nil {
		return nil, makeError(sErrors.VmNotFoundErr)
	}

	glc := gpu_lease_repo.New()

	lease, err := glc.WithVmID(vmID).Get()
	if err != nil {
		return nil, makeError(err)
	}

	if lease == nil {
		return nil, nil
	}

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

	if o.GpuGroupID != nil {
		// Specific GPU group's GPU leases are requested
		glc.WithGpuGroupID(*o.GpuGroupID)
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

	leases, err := glc.List()
	if err != nil {
		return nil, makeError(err)
	}

	return leases, nil
}

// Create creates a GPU lease
//
// The lease is not active immediately, but will be activated by the GPU lease worker
func (c *Client) Create(leaseID, userID string, dtoGpuLeaseCreate *body.GpuLeaseCreate) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create gpu lease. details: %w", err)
	}

	if c.V2.HasAuth() && !c.V2.Auth().IsAdmin {
		// If not admin, the user cannot lease a GPU forever
		dtoGpuLeaseCreate.LeaseForever = false
	}

	params := model.GpuLeaseCreateParams{}.FromDTO(dtoGpuLeaseCreate)

	var leaseDuration float64
	if !c.V2.HasAuth() || (params.LeaseForever && c.V2.Auth().IsAdmin) {
		leaseDuration = 1000 * 365 * 24 // A 1000-year lease is close enough to forever, right? :)
	}

	// Find the lease duration by the user's plan
	if c.V2.HasAuth() {
		leaseDuration = c.V2.Auth().GetEffectiveRole().Quotas.GpuLeaseDuration
	}

	if leaseDuration == 0 {
		return makeError(errors.New("lease duration could not be determined"))
	}

	err := gpu_lease_repo.New().Create(leaseID, userID, params.GpuGroupName, leaseDuration)
	if err != nil {
		if errors.Is(err, gpu_lease_repo.GpuLeaseAlreadyExistsErr) {
			return makeError(sErrors.GpuLeaseAlreadyExistsErr)
		}

		return makeError(err)
	}

	return nil
}

// Update updates a GPU lease
func (c *Client) Update(id string, dtoGpuLeaseUpdate *body.GpuLeaseUpdate) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update gpu lease. details: %w", err)
	}

	lease, err := c.Get(id)
	if err != nil {
		return makeError(err)
	}

	if lease == nil {
		return makeError(sErrors.GpuLeaseNotFoundErr)
	}

	params := model.GpuLeaseUpdateParams{}.FromDTO(dtoGpuLeaseUpdate)

	// Ensure we active the lease if it is the first time attaching a VM
	if params.VmID != nil && lease.ActivatedAt == nil {
		now := time.Now()
		params.ActivatedAt = &now
	}

	// If the lease was request to be activated, check if that is allowed but checking if the lease is assigned
	if params.ActivatedAt != nil && lease.AssignedAt == nil {
		return makeError(sErrors.GpuLeaseNotAssignedErr)
	}

	// If the lease is trying to update to the same VM, ignore the attach
	if lease.VmID != nil && params.VmID != nil && *lease.VmID == *params.VmID {
		params.VmID = nil
		params.ActivatedAt = nil
	}

	err = gpu_lease_repo.New().UpdateWithParams(id, params)
	if err != nil {
		return makeError(err)
	}

	// If the lease already has a VM attached, detach it
	if lease.VmID != nil {
		err = c.V2.VMs().K8s().DetachGPU(*lease.VmID)
		if err != nil && !errors.Is(err, sErrors.VmNotFoundErr) {
			return makeError(err)
		}
	}

	// Attach the GPU to the VM
	if params.VmID != nil {
		group, err := c.V2.VMs().GpuGroups().Get(lease.GpuGroupID)
		if err != nil {
			return makeError(err)
		}

		err = c.V2.VMs().K8s().AttachGPU(*params.VmID, group.Name)
		if err != nil {
			return makeError(err)
		}
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

	// Detach the GPU
	if lease.VmID != nil {
		err = c.V2.VMs().K8s().DetachGPU(*lease.VmID)
		if err != nil && !errors.Is(err, sErrors.VmNotFoundErr) {
			return makeError(err)
		}
	}

	err = gpu_lease_repo.New().DeleteByID(id)
	if err != nil {
		return makeError(err)
	}

	return nil
}

// Count counts the number of GPU leases
func (c *Client) Count(opts ...opts.ListGpuLeaseOpts) (int, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to count gpu leases. details: %w", err)
	}

	o := sUtils.GetFirstOrDefault(opts)

	glc := gpu_lease_repo.New()

	if o.GpuGroupID != nil {
		glc.WithGpuGroupID(*o.GpuGroupID)
	}

	if o.CreatedBefore != nil {
		glc.CreatedBefore(*o.CreatedBefore)
	}

	count, err := glc.Count()
	if err != nil {
		return 0, makeError(err)
	}

	return count, nil
}

// GetQueuePosition fetches the queue position of a GPU lease.
// Queue position is the number of leases that were created before this one minus the total GPUs of the group
// A queue position of 0 means the lease can be activated.
func (c *Client) GetQueuePosition(id string) (int, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get gpu lease context. details: %w", err)
	}

	lease, err := c.Get(id)
	if err != nil {
		return 0, makeError(err)
	}

	if lease == nil {
		return 0, makeError(sErrors.GpuLeaseNotFoundErr)
	}

	count, err := c.Count(opts.ListGpuLeaseOpts{
		GpuGroupID:    &lease.GpuGroupID,
		CreatedBefore: &lease.CreatedAt,
	})

	gpuGroup, err := c.V2.VMs().GpuGroups().Get(lease.GpuGroupID)
	if err != nil {
		return 0, makeError(err)
	}

	if gpuGroup == nil {
		return 0, makeError(sErrors.GpuGroupNotFoundErr)
	}

	// Add 1 to the queue position to make it human-readable (queue position 1 means next in line)
	queuePosition := max((count-gpuGroup.Total)+1, 0)

	return queuePosition, nil
}
