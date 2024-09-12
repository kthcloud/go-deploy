package harbor_service

import (
	configModels "github.com/kthcloud/go-deploy/models/config"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/subsystems/harbor"
	"github.com/kthcloud/go-deploy/service/core"
	sErrors "github.com/kthcloud/go-deploy/service/errors"
	"github.com/kthcloud/go-deploy/service/generators"
	"github.com/kthcloud/go-deploy/service/v2/deployments/client"
	"github.com/kthcloud/go-deploy/service/v2/deployments/opts"
	"github.com/kthcloud/go-deploy/service/v2/deployments/resources"
	"github.com/kthcloud/go-deploy/utils/subsystemutils"
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
func (c *Client) Get(opts *opts.Opts) (*model.Deployment, *harbor.Client, generators.HarborGenerator, error) {
	var deployment *model.Deployment
	var harborClient *harbor.Client
	var harborGenerator generators.HarborGenerator
	var err error

	if opts.DeploymentID != "" {
		deployment, err = c.Deployment(opts.DeploymentID, nil)
		if err != nil {
			return nil, nil, nil, err
		}

		if deployment == nil {
			return nil, nil, nil, sErrors.DeploymentNotFoundErr
		}
	}

	if opts.Client {
		var userID string
		if opts.ExtraOpts.UserID != "" {
			userID = opts.ExtraOpts.UserID
		} else {
			userID = deployment.OwnerID
		}

		harborClient, err = c.Client(userID)
		if err != nil {
			return nil, nil, nil, err
		}

		if harborClient == nil {
			return nil, nil, nil, sErrors.DeploymentNotFoundErr
		}
	}

	if opts.Generator {
		var userID string
		if opts.ExtraOpts.UserID != "" {
			userID = opts.ExtraOpts.UserID
		} else {
			userID = deployment.OwnerID
		}

		var zone *configModels.Zone
		if opts.ExtraOpts.Zone != nil {
			zone = opts.ExtraOpts.Zone
		} else if deployment != nil {
			zone = config.Config.GetZone(deployment.Zone)
		}

		harborGenerator = c.Generator(deployment, userID, zone)
		if harborGenerator == nil {
			return nil, nil, nil, sErrors.DeploymentNotFoundErr
		}
	}

	return deployment, harborClient, harborGenerator, nil
}

// Client returns the Harbor service client.
func (c *Client) Client(userID string) (*harbor.Client, error) {
	if userID == "" {
		panic("user id is empty")
	}

	return withClient(getProjectName(userID))
}

// Generator returns the Harbor generator.
func (c *Client) Generator(deployment *model.Deployment, userID string, zone *configModels.Zone) generators.HarborGenerator {
	if userID == "" {
		panic("user id is empty")
	}

	if zone == nil {
		panic("deployment zone is nil")
	}

	return resources.Harbor(deployment, zone, getProjectName(userID))
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
