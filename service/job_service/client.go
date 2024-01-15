package job_service

import (
	jobModels "go-deploy/models/sys/job"
	"go-deploy/service"
)

// Client is the client for the Job service.
type Client struct {
	// Cache is used to cache the resources fetched inside the service.
	Cache *service.Cache

	// Auth is the authentication information for the client.
	Auth *service.AuthInfo
}

// New creates a new job service client.
func New() *Client {
	return &Client{
		Cache: service.NewCache(),
	}
}

// WithAuth sets the auth on the context.
// This is used to perform authorization checks.
func (c *Client) WithAuth(auth *service.AuthInfo) *Client {
	c.Auth = auth
	return c
}

// Job returns the job with the given ID.
// After a successful fetch, the job will be cached.
func (c *Client) Job(id string, jmc *jobModels.Client) (*jobModels.Job, error) {
	job := c.Cache.GetJob(id)
	if job == nil {
		var err error
		job, err = jmc.GetByID(id)
		if err != nil {
			return nil, err
		}

		c.Cache.StoreJob(job)
	}

	return job, nil
}

// Jobs returns a list of jobs.
// After a successful fetch, the jobs will be cached.
func (c *Client) Jobs(jmc *jobModels.Client) ([]jobModels.Job, error) {
	// Right now we don't have a way to skip fetching when requesting a list of resources
	jobs, err := jmc.List()
	if err != nil {
		return nil, err
	}

	for _, job := range jobs {
		c.Cache.StoreJob(&job)
	}

	return jobs, nil
}

// RefreshJob clears the cache for the job with the given ID and fetches it again.
// After a successful fetch, the job will be cached.
func (c *Client) RefreshJob(id string, jmc *jobModels.Client) (*jobModels.Job, error) {
	job, err := jmc.GetByID(id)
	if err != nil {
		return nil, err
	}

	c.Cache.StoreJob(job)
	return job, nil
}
