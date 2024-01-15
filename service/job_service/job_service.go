package job_service

import (
	"go-deploy/models/dto/body"
	jobModels "go-deploy/models/sys/job"
	"go-deploy/service"
	sErrors "go-deploy/service/errors"
)

// Get retrieves a job by ID.
func (c *Client) Get(id string, opts ...*GetOpts) (*jobModels.Job, error) {
	_ = service.GetFirstOrDefault(opts)

	jmc := jobModels.New()

	if c.Auth != nil && !c.Auth.IsAdmin {
		jmc.RestrictToUser(c.Auth.UserID)
	}

	return c.Job(id, jmc)
}

// List retrieves a list of jobs.
func (c *Client) List(opt ...ListOpts) ([]jobModels.Job, error) {
	o := service.GetFirstOrDefault(opt)

	jmc := jobModels.New()

	if o.Pagination != nil {
		jmc.WithPagination(o.Pagination.Page, o.Pagination.PageSize)
	}

	var effectiveUserID string
	if o.UserID != nil {
		// Specific user's jobs are requested
		if c.Auth == nil || c.Auth.UserID == *o.UserID || c.Auth.IsAdmin {
			effectiveUserID = *o.UserID
		} else {
			// User cannot access the other user's resources
			effectiveUserID = c.Auth.UserID
		}
	} else {
		// All jobs are requested
		if c.Auth != nil && !c.Auth.IsAdmin {
			effectiveUserID = c.Auth.UserID
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
	if c.Auth != nil && !c.Auth.IsAdmin {
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
