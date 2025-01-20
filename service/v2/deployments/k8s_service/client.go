package k8s_service

import (
	configModels "github.com/kthcloud/go-deploy/models/config"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s"
	"github.com/kthcloud/go-deploy/service/core"
	sErrors "github.com/kthcloud/go-deploy/service/errors"
	"github.com/kthcloud/go-deploy/service/generators"
	"github.com/kthcloud/go-deploy/service/v2/deployments/client"
	"github.com/kthcloud/go-deploy/service/v2/deployments/opts"
	"github.com/kthcloud/go-deploy/service/v2/deployments/resources"
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

func OptsOnlyClient(zone *configModels.Zone) *opts.Opts {
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

// New creates a new Client and injects the cache.
// If a cache is not supplied it will create a new one.
func New(cache ...*core.Cache) *Client {
	var ca *core.Cache
	if len(cache) > 0 {
		ca = cache[0]
	} else {
		ca = core.NewCache()
	}

	c := &Client{BaseClient: client.NewBaseClient[Client](ca)}
	c.BaseClient.SetParent(c)
	return c
}

// Get returns the deployment, client, and generator.
//
// Depending on the options specified, some return values may be nil.
// This is useful when you don't always need all the resources.
func (c *Client) Get(opts *opts.Opts) (*model.Deployment, *k8s.Client, generators.K8sGenerator, error) {
	var deployment *model.Deployment
	var k8sClient *k8s.Client
	var k8sGenerator generators.K8sGenerator
	var err error

	if opts.DeploymentID != "" {
		deployment, err = c.Deployment(opts.DeploymentID, nil)
		if err != nil {
			return nil, nil, nil, err
		}

		if deployment == nil {
			return nil, nil, nil, sErrors.ErrDeploymentNotFound
		}
	}

	if opts.Client {
		zone := getZone(opts, deployment)
		if zone == nil {
			return nil, nil, nil, sErrors.ErrZoneNotFound
		}

		k8sClient, err = c.Client(zone)
		if err != nil {
			return nil, nil, nil, err
		}

		if k8sClient == nil {
			return nil, nil, nil, sErrors.ErrDeploymentNotFound
		}
	}

	if opts.Generator {
		zone := getZone(opts, deployment)
		if zone == nil {
			return nil, nil, nil, sErrors.ErrZoneNotFound
		}

		k8sGenerator = c.Generator(deployment, k8sClient, zone)
		if k8sGenerator == nil {
			return nil, nil, nil, sErrors.ErrDeploymentNotFound
		}
	}

	return deployment, k8sClient, k8sGenerator, nil
}

// Client returns the K8s service client.
func (c *Client) Client(zone *configModels.Zone) (*k8s.Client, error) {
	return withClient(zone, getNamespaceName(zone))
}

// Generator returns the K8s generator.
func (c *Client) Generator(deployment *model.Deployment, client *k8s.Client, zone *configModels.Zone) generators.K8sGenerator {
	if deployment == nil {
		panic("deployment is nil")
	}

	if client == nil {
		panic("client is nil")
	}

	if zone == nil {
		panic("deployment zone is nil")
	}

	return resources.K8s(deployment, zone, client, getNamespaceName(zone))
}

// getNamespaceName returns the namespace name.
func getNamespaceName(zone *configModels.Zone) string {
	return zone.K8s.Namespaces.Deployment
}

// withClient returns a new K8s service client.
func withClient(zone *configModels.Zone, namespace string) (*k8s.Client, error) {
	return k8s.New(&k8s.ClientConf{
		K8sClient:         zone.K8s.Client,
		KubeVirtK8sClient: zone.K8s.KubeVirtClient,
		Namespace:         namespace,
	})
}

// getZone is a helper function that returns either the zone in opts or the zone in the deployment.
func getZone(opts *opts.Opts, d *model.Deployment) *configModels.Zone {
	if opts.ExtraOpts.Zone != nil {
		return opts.ExtraOpts.Zone
	}

	if d != nil {
		return config.Config.GetZone(d.Zone)
	}

	return nil
}
