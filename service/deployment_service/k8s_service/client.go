package k8s_service

import (
	configModels "go-deploy/models/config"
	deploymentModels "go-deploy/models/sys/deployment"
	"go-deploy/pkg/config"
	"go-deploy/pkg/subsystems/k8s"
	"go-deploy/service"
	"go-deploy/service/deployment_service/client"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/resources"
	"go-deploy/utils/subsystemutils"
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

func OptsNoGenerator(deploymentID string, overwriteOps ...client.ExtraOpts) *client.Opts {
	var eo client.ExtraOpts
	if len(overwriteOps) > 0 {
		eo = overwriteOps[0]
	}
	return &client.Opts{
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
func New(cache *service.Cache) *Client {
	c := &Client{
		BaseClient: client.NewBaseClient[Client](cache),
	}
	c.BaseClient.SetParent(c)
	return c
}

// Get returns the deployment, client, and generator.
//
// Depending on the options specified, some return values may be nil.
// This is useful when you don't always need all the resources.
func (c *Client) Get(opts *client.Opts) (*deploymentModels.Deployment, *k8s.Client, *resources.K8sGenerator, error) {
	var d *deploymentModels.Deployment
	var kc *k8s.Client
	var g *resources.K8sGenerator
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

		var zone *configModels.DeploymentZone
		if opts.ExtraOpts.Zone != nil {
			zone = opts.ExtraOpts.Zone
		} else if d != nil {
			zone = config.Config.Deployment.GetZone(d.Zone)
		}

		kc, err = c.Client(userID, zone)
		if err != nil {
			return nil, nil, nil, err
		}

		if kc == nil {
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

		g = c.Generator(d, kc, zone)
		if g == nil {
			return nil, nil, nil, sErrors.DeploymentNotFoundErr
		}
	}

	return d, kc, g, nil
}

// Client returns the K8s service client.
func (c *Client) Client(userID string, zone *configModels.DeploymentZone) (*k8s.Client, error) {
	if userID == "" {
		panic("user id is empty")
	}

	return withClient(zone, getNamespaceName(userID))
}

// Generator returns the K8s generator.
func (c *Client) Generator(d *deploymentModels.Deployment, client *k8s.Client, zone *configModels.DeploymentZone) *resources.K8sGenerator {
	if d == nil {
		panic("deployment is nil")
	}

	if client == nil {
		panic("client is nil")
	}

	if zone == nil {
		panic("deployment zone is nil")
	}

	return resources.PublicGenerator().WithDeployment(d).WithDeploymentZone(zone).K8s(client)
}

// getNamespaceName returns the namespace name for the user.
func getNamespaceName(userID string) string {
	return subsystemutils.GetPrefixedName(userID)
}

// withClient returns a new K8s service client.
func withClient(zone *configModels.DeploymentZone, namespace string) (*k8s.Client, error) {
	return k8s.New(zone.Client, namespace)
}
