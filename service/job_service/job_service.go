package job_service

import (
	"go-deploy/models/dto/body"
	jobModel "go-deploy/models/sys/job"
	"go-deploy/service"
	sErrors "go-deploy/service/errors"
)

func (c *Client) Get(id string, opts ...*GetOpts) (*jobModel.Job, error) {
	_ = service.GetFirstOrDefault(opts)

	jmc := jobModel.New()

	if c.Auth != nil && !c.Auth.IsAdmin {
		jmc.RestrictToUser(c.Auth.UserID)
	}

	return c.Job(id, jmc)
}

func (c *Client) List(opt ...ListOpts) ([]jobModel.Job, error) {
	o := service.GetFirstOrDefault(opt)

	jmc := jobModel.New()

	if o.Pagination != nil {
		jmc.AddPagination(o.Pagination.Page, o.Pagination.PageSize)
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

func (c *Client) Create(id, userID, jobType string, args map[string]interface{}) error {
	return jobModel.New().Create(id, userID, jobType, args)
}

func (c *Client) Update(id string, jobUpdateDTO *body.JobUpdate) (*jobModel.Job, error) {
	if c.Auth != nil && !c.Auth.IsAdmin {
		return nil, sErrors.ForbiddenErr
	}

	var params jobModel.UpdateParams
	params.FromDTO(jobUpdateDTO)

	jmc := jobModel.New()

	err := jmc.UpdateWithParams(id, &params)
	if err != nil {
		return nil, err
	}

	return c.RefreshJob(id, jmc)
}

func (c *Client) Exists(id string) (bool, error) {
	return jobModel.New().ExistsByID(id)
}
