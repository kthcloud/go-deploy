package github_service

import (
	"go-deploy/models/sys/deployment"
	"go-deploy/pkg/subsystems/github"
	"go-deploy/service/deployment_service/client"
	dErrors "go-deploy/service/deployment_service/errors"
	"go-deploy/service/resources"
)

type Opts struct {
	Deployment bool
	Client     bool
	Generator  bool
}

var (
	OptsAll = &Opts{
		Deployment: true,
		Client:     true,
		Generator:  true,
	}
	OptsNoDeployment = &Opts{
		Deployment: false,
		Client:     true,
		Generator:  true,
	}
	OptsNoGenerator = &Opts{
		Deployment: true,
		Client:     true,
		Generator:  false,
	}
	OptsOnlyDeployment = &Opts{
		Deployment: true,
		Client:     false,
		Generator:  false,
	}
	OptsOnlyClient = &Opts{
		Deployment: false,
		Client:     true,
		Generator:  false,
	}
)

type Client struct {
	client.BaseClient[Client]

	client    *github.Client
	generator *resources.GitHubGenerator

	token        string
	repositoryID int64
	repository   *deployment.GitHubRepository
}

func New(context *client.Context) *Client {
	c := &Client{}
	c.BaseClient.SetParent(c)
	if context != nil {
		c.BaseClient.SetContext(context)
	}
	return c
}

func (c *Client) Get(opts *Opts) (*deployment.Deployment, *github.Client, *resources.GitHubGenerator, error) {
	var d *deployment.Deployment
	if opts.Deployment {
		d = c.Deployment()
		if d == nil {
			return nil, nil, nil, dErrors.DeploymentNotFoundErr
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
			return nil, nil, nil, dErrors.DeploymentNotFoundErr
		}
	}

	var g *resources.GitHubGenerator
	if opts.Generator {
		g = c.Generator()
		if g == nil {
			return nil, nil, nil, dErrors.DeploymentNotFoundErr
		}
	}

	return d, gc, g, nil
}

func (c *Client) Client() *github.Client {
	return c.client
}

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

func (c *Client) Token() string {
	return c.token
}

func (c *Client) RepositoryID() int64 {
	return c.repositoryID
}

func (c *Client) Repository() *deployment.GitHubRepository {
	return c.repository
}

func (c *Client) WithToken(token string) *Client {
	c.token = token
	return c
}

func (c *Client) WithRepositoryID(repositoryID int64) *Client {
	c.repositoryID = repositoryID
	return c
}

func (c *Client) WithRepository(repository *deployment.GitHubRepository) *Client {
	c.repository = repository
	return c
}
