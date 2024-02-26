package k8s_service

import (
	configModels "go-deploy/models/config"
	smModels "go-deploy/models/sys/sm"
	"go-deploy/pkg/config"
	"go-deploy/pkg/subsystems/k8s"
	"go-deploy/service/core"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/resources"
	"go-deploy/service/v1/sms/client"
	"go-deploy/service/v1/sms/opts"
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
	generator *resources.K8sGenerator
}

// New creates a new Client.
func New(cache *core.Cache) *Client {
	c := &Client{BaseClient: client.NewBaseClient[Client](cache)}
	c.BaseClient.SetParent(c)
	return c
}

// Get returns the deployment, client, and generator.
//
// Depending on the options specified, some return values may be nil.
// This is useful when you don't always need all the resources.
func (c *Client) Get(opts *opts.Opts) (*smModels.SM, *k8s.Client, *resources.K8sGenerator, error) {
	var sm *smModels.SM
	var kc *k8s.Client
	var g *resources.K8sGenerator
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

		kc, err = c.Client(userID, zone)
		if err != nil {
			return nil, nil, nil, err
		}

		if kc == nil {
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

		g = c.Generator(sm, kc, zone)
		if g == nil {
			return nil, nil, nil, sErrors.SmNotFoundErr
		}
	}

	return sm, kc, g, nil
}

// Client returns the K8s service client.
func (c *Client) Client(userID string, zone *configModels.DeploymentZone) (*k8s.Client, error) {
	if userID == "" {
		panic("user id is empty")
	}

	return withClient(zone, getNamespaceName(zone))
}

// Generator returns the K8s generator.
func (c *Client) Generator(sm *smModels.SM, client *k8s.Client, zone *configModels.DeploymentZone) *resources.K8sGenerator {
	if sm == nil {
		panic("deployment is nil")
	}

	if client == nil {
		panic("client is nil")
	}

	if zone == nil {
		panic("deployment zone is nil")
	}

	return resources.PublicGenerator().WithSM(sm).WithDeploymentZone(zone).K8s(client)
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
