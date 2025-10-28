package k8s_service

import (
	configModels "github.com/kthcloud/go-deploy/models/config"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s"
	"github.com/kthcloud/go-deploy/service/core"
	sErrors "github.com/kthcloud/go-deploy/service/errors"
	"github.com/kthcloud/go-deploy/service/generators"
	"github.com/kthcloud/go-deploy/service/v2/gpu_claims/client"
	"github.com/kthcloud/go-deploy/service/v2/gpu_claims/opts"
	"github.com/kthcloud/go-deploy/service/v2/gpu_claims/resources"
)

// OptsAll returns the options required to get all the service tools, ie. GPUClaim, client, and generator.
func OptsAll(claimID string, overwriteOps ...opts.ExtraOpts) *opts.Opts {
	var eo opts.ExtraOpts
	if len(overwriteOps) > 0 {
		eo = overwriteOps[0]
	}

	return &opts.Opts{
		ClaimID:   claimID,
		Client:    true,
		Generator: true,
		ExtraOpts: eo,
	}
}

// OptsNoGenerator returns the options required to get the SM and client.
func OptsNoGenerator(rcID string, overwriteOps ...opts.ExtraOpts) *opts.Opts {
	var eo opts.ExtraOpts
	if len(overwriteOps) > 0 {
		eo = overwriteOps[0]
	}
	return &opts.Opts{
		ClaimID:   rcID,
		Client:    true,
		ExtraOpts: eo,
	}
}

// Client is the client for the Harbor service.
// It contains a BaseClient, which is used to lazy-load and cache data.
type Client struct {
	client.BaseClient[Client]

	client    *k8s.Client
	generator *generators.K8sGenerator
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
func (c *Client) Get(opts *opts.Opts) (*model.GpuClaim, *k8s.Client, generators.K8sGenerator, error) {
	var gc *model.GpuClaim
	var k8sClient *k8s.Client
	var k8sGenerator generators.K8sGenerator
	var err error

	if opts.ClaimID != "" {
		gc, err = c.GpuClaim(opts.ClaimID, nil)
		if err != nil {
			return nil, nil, nil, err
		}

		if gc == nil {
			return nil, nil, nil, sErrors.ErrResourceNotFound
		}
	}

	if opts.Client {
		var zone *configModels.Zone
		if opts.ExtraOpts.Zone != nil {
			zone = opts.ExtraOpts.Zone
		} else if gc != nil {
			zone = config.Config.GetZone(gc.Zone)
		}

		if zone == nil {
			return nil, nil, nil, sErrors.ErrZoneNotFound
		}

		k8sClient, err = c.Client(zone)
		if err != nil {
			return nil, nil, nil, err
		}

		if k8sClient == nil {
			return nil, nil, nil, sErrors.ErrSmNotFound
		}
	}

	if opts.Generator {
		var zone *configModels.Zone
		if opts.ExtraOpts.Zone != nil {
			zone = opts.ExtraOpts.Zone
		} else if gc != nil {
			zone = config.Config.GetZone(gc.Zone)
		}

		k8sGenerator = c.Generator(gc, k8sClient, zone)
		if k8sGenerator == nil {
			return nil, nil, nil, sErrors.ErrSmNotFound
		}
	}

	return gc, k8sClient, k8sGenerator, nil
}

// Client returns the K8s service client.
func (c *Client) Client(zone *configModels.Zone) (*k8s.Client, error) {
	return withClient(zone, getNamespaceName(zone))
}

// Generator returns the K8s generator.
func (c *Client) Generator(gc *model.GpuClaim, client *k8s.Client, zone *configModels.Zone) generators.K8sGenerator {
	if gc == nil {
		panic("cpuclaim is nil")
	}

	if client == nil {
		panic("client is nil")
	}

	if zone == nil {
		panic("gpuclaim zone is nil")
	}

	return resources.K8s(gc, zone, client, getNamespaceName(zone))
}

// getNamespaceName returns the namespace name
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
