package gpu_groups

import (
	"fmt"
	"go-deploy/models/model"
	"go-deploy/pkg/db/resources/gpu_group_repo"
	sUtils "go-deploy/service/utils"
	"go-deploy/service/v2/vms/opts"
)

// Get gets a GPU group by ID
func (c *Client) Get(id string, opts ...opts.GetGpuGroupOpts) (*model.GpuGroup, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get gpu group. details: %w", err)
	}

	ggc := gpu_group_repo.New()

	group, err := ggc.GetByID(id)
	if err != nil {
		return nil, makeError(err)
	}

	return group, nil
}

// List lists GPU groups for a VM
func (c *Client) List(opts ...opts.ListGpuGroupOpts) ([]model.GpuGroup, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to list gpu groups. details: %w", err)
	}

	o := sUtils.GetFirstOrDefault(opts)

	ggc := gpu_group_repo.New()

	if o.Pagination != nil {
		ggc.WithPagination(o.Pagination.Page, o.Pagination.PageSize)
	}

	groups, err := ggc.List()
	if err != nil {
		return nil, makeError(err)
	}

	return groups, nil
}

// Exists checks if a GPU group exists
func (c *Client) Exists(id string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if gpu group exists. details: %w", err)
	}

	ggc := gpu_group_repo.New()

	exists, err := ggc.ExistsByID(id)
	if err != nil {
		return false, makeError(err)
	}

	return exists, nil
}
