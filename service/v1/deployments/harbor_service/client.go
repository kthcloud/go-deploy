package harbor_service

import (
	configModels "go-deploy/models/config"
	deploymentModels "go-deploy/models/sys/deployment"
	"go-deploy/pkg/config"
	"go-deploy/pkg/subsystems/harbor"
	"go-deploy/service/core"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/resources"
	"go-deploy/service/v1/deployments/client"
	"go-deploy/service/v1/deployments/opts"
	"go-deploy/utils/subsystemutils"
)

// OptsAll returns the options required to get all the service tools, ie. deployment, client, and generator.
func OptsAll(deploymentID string, overwriteOps ...opts.ExtraOpts) *opts.Opts {
	var eo opts.ExtraOpts
	if len(overwriteOps) > 0 {
		eo = overwriteOps[0]
	}

	return &opts.Opts{
		DeploymentID: deploymentID,
		Client:       true,
		Generator:    true,
		ExtraOpts:    eo,
	}
}

// OptsNoGenerator returns the options required to get all the service tools, ie. deployment, client, and generator.
func OptsNoGenerator(deploymentID string, overwriteOps ...opts.ExtraOpts) *opts.Opts {
	var eo opts.ExtraOpts
	if len(overwriteOps) > 0 {
		eo = overwriteOps[0]
	}
	return &opts.Opts{
		DeploymentID: deploymentID,
		Client:       true,
		ExtraOpts:    eo,
	}
}

// Client is the client for the Harbor service.
// It contains a BaseClient, which is used to lazy-load and cache data.
type Client struct {
	client.BaseClient[Client]
}

// New creates a new Client.
// If context is not nil, it will be used to create a new Client.
// Otherwise, an empty context will be created.
func New(cache *core.Cache) *Client {
	c := &Client{BaseClient: client.NewBaseClient[Client](cache)}
	c.BaseClient.SetParent(c)
	return c
}

// Get returns the deployment, client, and generator.
//
// Depending on the options specified, some return values may be nil.
// This is useful when you don't always need all the resources.
func (c *Client) Get(opts *opts.Opts) (*deploymentModels.Deployment, *harbor.Client, *resources.HarborGenerator, error) {
	var d *deploymentModels.Deployment
	var gc *harbor.Client
	var g *resources.HarborGenerator
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
		var userID string
		if opts.ExtraOpts.UserID != "" {
			userID = opts.ExtraOpts.UserID
		} else {
			userID = d.OwnerID
		}

		gc, err = c.Client(userID)
		if err != nil {
			return nil, nil, nil, err
		}

		if gc == nil {
			return nil, nil, nil, sErrors.DeploymentNotFoundErr
		}
	}

	if opts.Generator {
		var userID string
		if opts.ExtraOpts.UserID != "" {
			userID = opts.ExtraOpts.UserID
		} else {
			userID = d.OwnerID
		}

		var zone *configModels.DeploymentZone
		if opts.ExtraOpts.Zone != nil {
			zone = opts.ExtraOpts.Zone
		} else if d != nil {
			zone = config.Config.Deployment.GetZone(d.Zone)
		}

		g = c.Generator(d, userID, zone)
		if g == nil {
			return nil, nil, nil, sErrors.DeploymentNotFoundErr
		}
	}

	return d, gc, g, nil
}

// Client returns the Harbor service client.
func (c *Client) Client(userID string) (*harbor.Client, error) {
	if userID == "" {
		panic("user id is empty")
	}

	return withClient(getProjectName(userID))
}

// Generator returns the Harbor generator.
func (c *Client) Generator(d *deploymentModels.Deployment, userID string, zone *configModels.DeploymentZone) *resources.HarborGenerator {
	if userID == "" {
		panic("user id is empty")
	}

	if zone == nil {
		panic("deployment zone is nil")
	}

	return resources.PublicGenerator().WithDeploymentZone(zone).WithDeployment(d).Harbor(getProjectName(userID))
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
