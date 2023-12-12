package github_service

import (
	configModels "go-deploy/models/config"
	deploymentModels "go-deploy/models/sys/deployment"
	"go-deploy/pkg/config"
	"go-deploy/pkg/subsystems/github"
	"go-deploy/service/deployment_service/client"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/resources"
)

func OptsAll(deploymentID string, overwriteOps ...client.ExtraOpts) *client.Opts {
	var eo client.ExtraOpts
	if len(overwriteOps) > 0 {
		eo = overwriteOps[0]
	}

	return &client.Opts{
		DeploymentID: deploymentID,
		Client:       true,
		Generator:    true,
		ExtraOpts:    eo,
	}
}

func OptsOnlyDeployment(deploymentID string) *client.Opts {
	return &client.Opts{
		DeploymentID: deploymentID,
	}
}

func OptsOnlyClient() *client.Opts {
	return &client.Opts{
		Client: true,
	}
}

// Client is the client for the GitHub service.
// It contains a BaseClient, which is used to lazy-load and cache data.
type Client struct {
	client.BaseClient[Client]

	token        string
	repositoryID int64
}

// New creates a new Client.
// If context is not nil, it will be used to create a new BaseClient.
// Otherwise, an empty context will be created.
func New(context *client.Context) *Client {
	c := &Client{
		BaseClient: client.NewBaseClient[Client](context),
	}
	c.BaseClient.SetParent(c)
	return c
}

// Get returns the deployment, client, and generator.
//
// Depending on the options specified, some return values may be nil.
// This is useful when you don't always need all the resources.
func (c *Client) Get(opts *client.Opts) (*deploymentModels.Deployment, *github.Client, *resources.GitHubGenerator, error) {
	var d *deploymentModels.Deployment
	var gc *github.Client
	var g *resources.GitHubGenerator
	var err error

	if opts.DeploymentID != "" {
		d, err = c.Deployment(opts.DeploymentID, nil)
		if err != nil {
			return nil, nil, nil, err
		}

		if d == nil {
			return nil, nil, nil, sErrors.DeploymentNotFoundErr
		}
	}

	if opts.Client {
		gc, err = c.Client()
		if err != nil {
			return nil, nil, nil, err
		}

		if gc == nil {
			return nil, nil, nil, sErrors.DeploymentNotFoundErr
		}
	}

	if opts.Generator {
		var zone *configModels.DeploymentZone
		if opts.ExtraOpts.Zone != nil {
			zone = opts.ExtraOpts.Zone
		} else if d != nil {
			zone = config.Config.Deployment.GetZone(d.Zone)
		}

		g = c.Generator(d, zone)
		if g == nil {
			return nil, nil, nil, sErrors.DeploymentNotFoundErr
		}
	}

	return d, gc, g, nil
}

// Client returns the GitHub service client.
//
// WithToken must be called before this function.
func (c *Client) Client() (*github.Client, error) {
	if c.token == "" {
		panic("github token is nil")
	}

	gc, err := withClient(c.token)
	if err != nil {
		return nil, err
	}

	return gc, nil
}

// Generator returns the GitHub generator.
func (c *Client) Generator(d *deploymentModels.Deployment, zone *configModels.DeploymentZone) *resources.GitHubGenerator {
	var dZone *configModels.DeploymentZone
	if zone != nil {
		dZone = zone
	} else if d != nil {
		dZone = config.Config.Deployment.GetZone(d.Zone)
	}

	if dZone == nil {
		panic("deployment zone is nil")
	}

	if c.token == "" {
		panic("github token is not set")
	}

	if c.repositoryID == 0 {
		panic("github repository id is not set")
	}

	return resources.PublicGenerator().WithDeployment(d).WithDeploymentZone(dZone).GitHub(c.token, c.repositoryID)
}

// WithToken sets the GitHub token.
func (c *Client) WithToken(token string) *Client {
	c.token = token
	return c
}

func (c *Client) WithRepositoryID(repositoryID int64) *Client {
	c.repositoryID = repositoryID
	return c
}

// WithDeployment sets the deployment.
func withClient(token string) (*github.Client, error) {
	return github.New(&github.ClientConf{
		Token: token,
	})
}
