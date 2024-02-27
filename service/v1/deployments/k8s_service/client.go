package k8s_service

import (
	configModels "go-deploy/models/config"
	deploymentModels "go-deploy/models/sys/deployment"
	"go-deploy/pkg/config"
	"go-deploy/pkg/subsystems/k8s"
	"go-deploy/service/core"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/resources"
	"go-deploy/service/v1/deployments/client"
	"go-deploy/service/v1/deployments/opts"
)

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

func OptsOnlyClient(zone *configModels.DeploymentZone) *opts.Opts {
	return &opts.Opts{
		Client:    true,
		ExtraOpts: opts.ExtraOpts{Zone: zone},
	}
}

// Client is the client for the Harbor service.
// It contains a BaseClient, which is used to lazy-load and cache data.
type Client struct {
	client.BaseClient[Client]
}

// New creates a new Client.
func New(cache *core.Cache) *Client {
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
func (c *Client) Get(opts *opts.Opts) (*deploymentModels.Deployment, *k8s.Client, *resources.K8sGenerator, error) {
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
		zone := getZone(opts, d)
		if zone == nil {
			return nil, nil, nil, sErrors.ZoneNotFoundErr
		}

		kc, err = c.Client(zone)
		if err != nil {
			return nil, nil, nil, err
		}

		if kc == nil {
			return nil, nil, nil, sErrors.DeploymentNotFoundErr
		}
	}

	if opts.Generator {
		zone := getZone(opts, d)
		if zone == nil {
			return nil, nil, nil, sErrors.ZoneNotFoundErr
		}

		g = c.Generator(d, kc, zone)
		if g == nil {
			return nil, nil, nil, sErrors.DeploymentNotFoundErr
		}
	}

	return d, kc, g, nil
}

// Client returns the K8s service client.
func (c *Client) Client(zone *configModels.DeploymentZone) (*k8s.Client, error) {
	return withClient(zone, getNamespaceName(zone))
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

// getNamespaceName returns the namespace name.
func getNamespaceName(zone *configModels.DeploymentZone) string {
	return zone.Namespaces.Deployment
}

// withClient returns a new K8s service client.
func withClient(zone *configModels.DeploymentZone, namespace string) (*k8s.Client, error) {
	return k8s.New(&k8s.ClientConf{
		K8sClient:         zone.K8sClient,
		KubeVirtK8sClient: zone.KubeVirtClient,
		Namespace:         namespace,
	})
}

// getZone is a helper function that returns either the zone in opts or the zone in the deployment.
func getZone(opts *opts.Opts, d *deploymentModels.Deployment) *configModels.DeploymentZone {
	if opts.ExtraOpts.Zone != nil {
		return opts.ExtraOpts.Zone
	}

	if d != nil {
		return config.Config.Deployment.GetZone(d.Zone)
	}

	return nil
}
