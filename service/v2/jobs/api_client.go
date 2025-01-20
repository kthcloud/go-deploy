package jobs

import (
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/db/resources/job_repo"
	"github.com/kthcloud/go-deploy/service/clients"
	"github.com/kthcloud/go-deploy/service/core"
)

// Client is the client for the Job service.
type Client struct {
	// V2 is a reference to the parent client.
	V2 clients.V2

	// Cache is used to cache the resources fetched inside the service.
	Cache *core.Cache
}

// New creates a new job service client.
func New(v2 clients.V2, cache ...*core.Cache) *Client {
	var c *core.Cache
	if len(cache) > 0 {
		c = cache[0]
	} else {
		c = core.NewCache()
	}

	return &Client{
		V2:    v2,
		Cache: c,
	}
}

// Job returns the job with the given ID.
// After a successful fetch, the job will be cached.
func (c *Client) Job(id string, jmc *job_repo.Client) (*model.Job, error) {
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
func (c *Client) Jobs(jmc *job_repo.Client) ([]model.Job, error) {
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
func (c *Client) RefreshJob(id string, jmc *job_repo.Client) (*model.Job, error) {
	job, err := jmc.GetByID(id)
	if err != nil {
		return nil, err
	}

	c.Cache.StoreJob(job)
	return job, nil
}
