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
)

type Client struct {
	client.BaseClient[Client]

	client    *harbor.Client
	generator *resources.HarborGenerator
}

func New(context *client.Context) *Client {
	c := &Client{}
	c.BaseClient.SetParent(c)
	if context != nil {
		c.BaseClient.SetContext(context)
	}
	return c
}

func (c *Client) Get(opts *Opts) (*deployment.Deployment, *harbor.Client, *resources.HarborGenerator, error) {
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

func (c *Client) Client() *harbor.Client {
	return c.client
}

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

func getProjectName(userID string) string {
	return subsystemutils.GetPrefixedName(userID)
}

func withClient(project string) (*harbor.Client, error) {
	return harbor.New(&harbor.ClientConf{
		URL:      config.Config.Harbor.URL,
		Username: config.Config.Harbor.User,
		Password: config.Config.Harbor.Password,
		Project:  project,
	})
}
