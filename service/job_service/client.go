package job_service

import (
	jobModels "go-deploy/models/sys/job"
	"go-deploy/service"
)

type Client struct {
	// Cache is used to cache the resources fetched inside the service.
	Cache *service.Cache

	// Auth is the authentication information for the client.
	Auth *service.AuthInfo
}

func New() *Client {
	return &Client{
		Cache: service.NewCache(),
	}
}

// WithAuth sets the auth on the context.
func (c *Client) WithAuth(auth *service.AuthInfo) *Client {
	c.Auth = auth
	return c
}

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

func (c *Client) RefreshJob(id string, jmc *jobModels.Client) (*jobModels.Job, error) {
	job, err := jmc.GetByID(id)
	if err != nil {
		return nil, err
	}

	c.Cache.StoreJob(job)
	return job, nil
}
