package k8s_service

import (
	configModels "go-deploy/models/config"
	"go-deploy/models/model"
	"go-deploy/pkg/config"
	"go-deploy/pkg/subsystems/k8s"
	"go-deploy/service/core"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/generators"
	"go-deploy/service/v1/sms/client"
	"go-deploy/service/v1/sms/opts"
	"go-deploy/service/v1/sms/resources"
)

// OptsAll returns the options required to get all the service tools, ie. SM, client, and generator.
func OptsAll(smID string, overwriteOps ...opts.ExtraOpts) *opts.Opts {
	var eo opts.ExtraOpts
	if len(overwriteOps) > 0 {
		eo = overwriteOps[0]
	}

	return &opts.Opts{
		SmID:      smID,
		Client:    true,
		Generator: true,
		ExtraOpts: eo,
	}
}

// OptsNoGenerator returns the options required to get the SM and client.
func OptsNoGenerator(smID string, overwriteOps ...opts.ExtraOpts) *opts.Opts {
	var eo opts.ExtraOpts
	if len(overwriteOps) > 0 {
		eo = overwriteOps[0]
	}
	return &opts.Opts{
		SmID:      smID,
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
func (c *Client) Get(opts *opts.Opts) (*model.SM, *k8s.Client, generators.K8sGenerator, error) {
	var sm *model.SM
	var k8sClient *k8s.Client
	var k8sGenerator generators.K8sGenerator
	var err error

	if opts.SmID != "" {
		sm, err = c.SM(opts.SmID, "", nil)
		if err != nil {
			return nil, nil, nil, err
		}

		if sm == nil {
			return nil, nil, nil, sErrors.SmNotFoundErr
		}
	}

	if opts.Client {
		var userID string
		if opts.ExtraOpts.UserID != "" {
			userID = opts.ExtraOpts.UserID
		} else {
			userID = sm.OwnerID
		}

		var zone *configModels.DeploymentZone
		if opts.ExtraOpts.Zone != nil {
			zone = opts.ExtraOpts.Zone
		} else if sm != nil {
			zone = config.Config.Deployment.GetZone(sm.Zone)
		}

		k8sClient, err = c.Client(userID, zone)
		if err != nil {
			return nil, nil, nil, err
		}

		if k8sClient == nil {
			return nil, nil, nil, sErrors.SmNotFoundErr
		}
	}

	if opts.Generator {
		var zone *configModels.DeploymentZone
		if opts.ExtraOpts.Zone != nil {
			zone = opts.ExtraOpts.Zone
		} else if sm != nil {
			zone = config.Config.Deployment.GetZone(sm.Zone)
		}

		k8sGenerator = c.Generator(sm, k8sClient, zone)
		if k8sGenerator == nil {
			return nil, nil, nil, sErrors.SmNotFoundErr
		}
	}

	return sm, k8sClient, k8sGenerator, nil
}

// Client returns the K8s service client.
func (c *Client) Client(userID string, zone *configModels.DeploymentZone) (*k8s.Client, error) {
	if userID == "" {
		panic("user id is empty")
	}

	return withClient(zone, getNamespaceName(zone))
}

// Generator returns the K8s generator.
func (c *Client) Generator(sm *model.SM, client *k8s.Client, zone *configModels.DeploymentZone) generators.K8sGenerator {
	if sm == nil {
		panic("deployment is nil")
	}

	if client == nil {
		panic("client is nil")
	}

	if zone == nil {
		panic("deployment zone is nil")
	}

	return resources.K8s().WithClient(client).WithSM(sm).WithZone(zone)
}

// getNamespaceName returns the namespace name
func getNamespaceName(zone *configModels.DeploymentZone) string {
	return zone.Namespaces.System
}

// withClient returns a new K8s service client.
func withClient(zone *configModels.DeploymentZone, namespace string) (*k8s.Client, error) {
	return k8s.New(&k8s.ClientConf{
		K8sClient:         zone.K8sClient,
		KubeVirtK8sClient: zone.KubeVirtClient,
		Namespace:         namespace,
	})
}
