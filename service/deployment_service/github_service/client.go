package github_service

import (
	"go-deploy/models/sys/deployment"
	"go-deploy/pkg/subsystems/github"
	"go-deploy/service/deployment_service/client"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/resources"
)

// Client is the client for the GitHub service.
// It contains a BaseClient, which is used to lazy-load and cache data.
type Client struct {
	client.BaseClient[Client]

	client    *github.Client
	generator *resources.GitHubGenerator

	token        string
	repositoryID int64
	repository   *deployment.GitHubRepository
}

// New creates a new Client.
// If context is not nil, it will be used to create a new BaseClient.
// Otherwise, an empty context will be created.
func New(context *client.Context) *Client {
	c := &Client{}
	c.BaseClient.SetParent(c)
	c.BaseClient.SetContext(context)
	return c
}

// Get returns the deployment, client, and generator.
//
// Depending on the options specified, some return values may be nil.
// This is useful when you don't always need all the resources.
func (c *Client) Get(opts *client.Opts) (*deployment.Deployment, *github.Client, *resources.GitHubGenerator, error) {
	var d *deployment.Deployment
	if opts.Deployment {
		d = c.Deployment()
		if d == nil {
			return nil, nil, nil, sErrors.DeploymentNotFoundErr
		}
	}

	var gc *github.Client
	if opts.Client {
		var err error
		gc, err = c.GetOrCreateClient()
		if err != nil {
			return nil, nil, nil, err
		}

		if gc == nil {
			return nil, nil, nil, sErrors.DeploymentNotFoundErr
		}
	}

	var g *resources.GitHubGenerator
	if opts.Generator {
		g = c.Generator()
		if g == nil {
			return nil, nil, nil, sErrors.DeploymentNotFoundErr
		}
	}

	return d, gc, g, nil
}

// Client returns the GitHub service client.
//
// This does not create a new client if it does not exist.
func (c *Client) Client() *github.Client {
	return c.client
}

// GetOrCreateClient returns the GitHub service client.
//
// If the client does not exist, it will be created.
func (c *Client) GetOrCreateClient() (*github.Client, error) {
	if c.client == nil {
		if c.token == "" {
			panic("token is empty")
		}

		hc, err := withClient(c.token)
		if err != nil {
			return nil, err
		}

		c.client = hc
	}

	return c.client, nil
}

// Generator returns the GitHub generator.
//
// If the generator does not exist, it will be created.
// If creating a new generator, the current deployment and zone will be used.
// Set the deployment and zone before calling this function by using WithDeployment and WithZone.
func (c *Client) Generator() *resources.GitHubGenerator {
	if c.generator == nil {
		pg := resources.PublicGenerator()

		if c.Deployment() != nil {
			pg.WithDeployment(c.Deployment())
		}

		if c.Zone() != nil {
			pg.WithDeploymentZone(c.Zone())
		}

		c.generator = pg.GitHub(c.token, c.repositoryID)
	}

	return c.generator
}

// Token returns the GitHub token.
func (c *Client) Token() string {
	return c.token
}

// RepositoryID returns the GitHub repository ID.
func (c *Client) RepositoryID() int64 {
	return c.repositoryID
}

// Repository returns the GitHub repository.
func (c *Client) Repository() *deployment.GitHubRepository {
	return c.repository
}

// WithToken sets the GitHub token.
func (c *Client) WithToken(token string) *Client {
	c.token = token
	return c
}

// WithRepositoryID sets the GitHub repository ID.
func (c *Client) WithRepositoryID(repositoryID int64) *Client {
	c.repositoryID = repositoryID
	return c
}

// WithRepository sets the GitHub repository.
func (c *Client) WithRepository(repository *deployment.GitHubRepository) *Client {
	c.repository = repository
	return c
}

// WithDeployment sets the deployment.
func withClient(token string) (*github.Client, error) {
	return github.New(&github.ClientConf{
		Token: token,
	})
}
