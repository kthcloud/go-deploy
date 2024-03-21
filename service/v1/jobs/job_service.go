package jobs

import (
	"go-deploy/dto/v1/body"
	"go-deploy/models/model"
	"go-deploy/pkg/db/resources/job_repo"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/utils"
	"go-deploy/service/v1/jobs/opts"
)

// Get retrieves a job by ID.
func (c *Client) Get(id string, opts ...opts.GetOpts) (*model.Job, error) {
	_ = utils.GetFirstOrDefault(opts)

	jmc := job_repo.New()

	if c.V1.Auth() != nil && !c.V1.Auth().IsAdmin {
		jmc.RestrictToUser(c.V1.Auth().UserID)
	}

	return c.Job(id, jmc)
}

// List retrieves a list of jobs.
func (c *Client) List(opt ...opts.ListOpts) ([]model.Job, error) {
	o := utils.GetFirstOrDefault(opt)

	jmc := job_repo.New()

	if o.Pagination != nil {
		jmc.WithPagination(o.Pagination.Page, o.Pagination.PageSize)
	}

	if o.SortBy != nil {
		jmc.WithSort(o.SortBy.Field, o.SortBy.Order)
	}

	var effectiveUserID string
	if o.UserID != nil {
		// Specific user's jobs are requested
		if !c.V1.HasAuth() || c.V1.Auth().UserID == *o.UserID || c.V1.Auth().IsAdmin {
			effectiveUserID = *o.UserID
		} else {
			// User cannot access the other user's resources
			effectiveUserID = c.V1.Auth().UserID
		}
	} else {
		// All jobs are requested
		if c.V1.Auth() != nil && !c.V1.Auth().IsAdmin {
			effectiveUserID = c.V1.Auth().UserID
		}
	}

	if effectiveUserID != "" {
		jmc.RestrictToUser(effectiveUserID)
	}

	if len(o.JobTypes) > 0 {
		jmc.IncludeTypes(o.JobTypes...)
	}

	if len(o.ExcludeJobTypes) > 0 {
		jmc.ExcludeTypes(o.ExcludeJobTypes...)
	}

	if len(o.Status) > 0 {
		jmc.IncludeStatus(o.Status...)
	}

	if len(o.ExcludeStatus) > 0 {
		jmc.ExcludeStatus(o.ExcludeStatus...)
	}

	return c.Jobs(jmc)
}

// Create creates a new job.
func (c *Client) Create(id, userID, jobType, version string, args map[string]interface{}) error {
	return job_repo.New().Create(id, userID, jobType, version, args)
}

// Update updates a job.
func (c *Client) Update(id string, jobUpdateDTO *body.JobUpdate) (*model.Job, error) {
	if c.V1.Auth() != nil && !c.V1.Auth().IsAdmin {
		return nil, sErrors.ForbiddenErr
	}

	var params model.JobUpdateParams
	params.FromDTO(jobUpdateDTO)

	jmc := job_repo.New()

	err := jmc.UpdateWithParams(id, &params)
	if err != nil {
		return nil, err
	}

	return c.RefreshJob(id, jmc)
}

// Exists checks if a job exists.
func (c *Client) Exists(id string) (bool, error) {
	return job_repo.New().ExistsByID(id)
}
