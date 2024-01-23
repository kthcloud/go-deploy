package jobs

import (
	"go-deploy/models/dto/v1/body"
	jobModels "go-deploy/models/sys/job"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/utils"
	"go-deploy/service/v1/jobs/opts"
)

// Get retrieves a job by ID.
func (c *Client) Get(id string, opts ...opts.GetOpts) (*jobModels.Job, error) {
	_ = utils.GetFirstOrDefault(opts)

	jmc := jobModels.New()

	if c.V1.Auth() != nil && !c.V1.Auth().IsAdmin {
		jmc.RestrictToUser(c.V1.Auth().UserID)
	}

	return c.Job(id, jmc)
}

// List retrieves a list of jobs.
func (c *Client) List(opt ...opts.ListOpts) ([]jobModels.Job, error) {
	o := utils.GetFirstOrDefault(opt)

	jmc := jobModels.New()

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

	if o.JobType != nil {
		jmc.IncludeTypes(*o.JobType)
	}

	if o.Status != nil {
		jmc.IncludeStatus(*o.Status)
	}

	return c.Jobs(jmc)
}

// Create creates a new job.
func (c *Client) Create(id, userID, jobType string, args map[string]interface{}) error {
	return jobModels.New().Create(id, userID, jobType, args)
}

// Update updates a job.
func (c *Client) Update(id string, jobUpdateDTO *body.JobUpdate) (*jobModels.Job, error) {
	if c.V1.Auth() != nil && !c.V1.Auth().IsAdmin {
		return nil, sErrors.ForbiddenErr
	}

	var params jobModels.UpdateParams
	params.FromDTO(jobUpdateDTO)

	jmc := jobModels.New()

	err := jmc.UpdateWithParams(id, &params)
	if err != nil {
		return nil, err
	}

	return c.RefreshJob(id, jmc)
}

// Exists checks if a job exists.
func (c *Client) Exists(id string) (bool, error) {
	return jobModels.New().ExistsByID(id)
}