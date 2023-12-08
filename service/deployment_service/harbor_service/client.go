package harbor_service

import (
	"go-deploy/models/sys/deployment"
	"go-deploy/pkg/config"
	"go-deploy/pkg/subsystems/harbor"
	"go-deploy/service/deployment_service/client"
	dErrors "go-deploy/service/deployment_service/errors"
	"go-deploy/service/resources"
	"go-deploy/utils/subsystemutils"
)

// Client is the client for the Harbor service.
// It contains a BaseClient, which is used to lazy-load and cache data.
type Client struct {
	client.BaseClient[Client]

	client    *harbor.Client
	generator *resources.HarborGenerator
}

// New creates a new Client.
// If context is not nil, it will be used to create a new BaseClient.
// Otherwise, an empty context will be created.
func New(context *client.Context) *Client {
	c := &Client{}
	c.BaseClient.SetParent(c)
	if context != nil {
		c.BaseClient.SetContext(context)
	}
	return c
}

// Get returns the deployment, client, and generator.
//
// Depending on the options specified, some return values may be nil.
// This is useful when you don't always need all the resources.
func (c *Client) Get(opts *client.Opts) (*deployment.Deployment, *harbor.Client, *resources.HarborGenerator, error) {
	var d *deployment.Deployment
	if opts.Deployment {
		d = c.Deployment()
		if d == nil {
			return nil, nil, nil, dErrors.DeploymentNotFoundErr
		}
	}

	var hc *harbor.Client
	if opts.Client {
		var err error
		hc, err = c.GetOrCreateClient()
		if err != nil {
			return nil, nil, nil, err
		}

		if hc == nil {
			return nil, nil, nil, dErrors.DeploymentNotFoundErr
		}
	}

	var g *resources.HarborGenerator
	if opts.Generator {
		g = c.Generator()
		if g == nil {
			return nil, nil, nil, dErrors.DeploymentNotFoundErr
		}
	}

	return d, hc, g, nil
}

// Client returns the GitHub service client.
//
// This does not create a new client if it does not exist.
func (c *Client) Client() *harbor.Client {
	return c.client
}

// GetOrCreateClient returns the GitHub service client.
//
// If the client does not exist, it will be created.
func (c *Client) GetOrCreateClient() (*harbor.Client, error) {
	if c.client == nil {
		if c.UserID == "" {
			panic("user id is empty")
		}

		hc, err := withClient(getProjectName(c.UserID))
		if err != nil {
			return nil, err
		}

		c.client = hc
	}

	return c.client, nil
}

// WithUserID sets the user id
// Overwrites the base client's user id function
// This is used to set the project
func (c *Client) WithUserID(userID string) *Client {
	c.BaseClient.WithUserID(userID)

	hc := c.Client()
	if hc != nil {
		hc.Project = getProjectName(userID)
	}

	g := c.Generator()
	if g != nil {
		g = g.Harbor(getProjectName(userID))
	}

	return c
}

// Generator returns the GitHub generator.
//
// If the generator does not exist, it will be created.
// If creating a new generator, the current deployment and zone will be used.
// Set the deployment and zone before calling this function by using WithDeployment and WithZone.
func (c *Client) Generator() *resources.HarborGenerator {
	if c.generator == nil {
		pg := resources.PublicGenerator()

		if c.Deployment() != nil {
			pg.WithDeployment(c.Deployment())
		}

		if c.Zone() != nil {
			pg.WithDeploymentZone(c.Zone())
		}

		c.generator = pg.Harbor(getProjectName(c.UserID))
	}

	return c.generator
}

// getProjectName returns the project name for the user.
func getProjectName(userID string) string {
	return subsystemutils.GetPrefixedName(userID)
}

// withClient creates a new Harbor client.
func withClient(project string) (*harbor.Client, error) {
	return harbor.New(&harbor.ClientConf{
		URL:      config.Config.Harbor.URL,
		Username: config.Config.Harbor.User,
		Password: config.Config.Harbor.Password,
		Project:  project,
	})
}
