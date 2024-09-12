package k8s_service

import (
	"fmt"
	configModels "github.com/kthcloud/go-deploy/models/config"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/subsystems/k8s"
	"github.com/kthcloud/go-deploy/service/core"
	sErrors "github.com/kthcloud/go-deploy/service/errors"
	"github.com/kthcloud/go-deploy/service/generators"
	"github.com/kthcloud/go-deploy/service/v2/vms/client"
	"github.com/kthcloud/go-deploy/service/v2/vms/opts"
	"github.com/kthcloud/go-deploy/service/v2/vms/resources"
)

// OptsAll returns the options required to get all the service tools, ie. VM, client and generator.
func OptsAll(vmID string, overwriteOps ...opts.ExtraOpts) *opts.Opts {
	var ow opts.ExtraOpts
	if len(overwriteOps) > 0 {
		ow = overwriteOps[0]
	}

	return &opts.Opts{
		VmID:      vmID,
		Client:    true,
		Generator: true,
		ExtraOpts: ow,
	}
}

// OptsNoGenerator returns the options required to get only the VM and client.
func OptsNoGenerator(vmID string, extraOpts ...opts.ExtraOpts) *opts.Opts {
	var eo opts.ExtraOpts
	if len(extraOpts) > 0 {
		eo = extraOpts[0]
	}

	return &opts.Opts{
		VmID:      vmID,
		Client:    true,
		ExtraOpts: eo,
	}
}

// OptsOnlyClient returns the options required to get only the client.
func OptsOnlyClient(zone string, extraOpts ...opts.ExtraOpts) *opts.Opts {
	var eo opts.ExtraOpts
	eo.Zone = config.Config.GetZone(zone)
	if len(extraOpts) > 0 {
		eo = extraOpts[0]
	}

	return &opts.Opts{
		Client:    true,
		ExtraOpts: eo,
	}
}

// Client is the client for the Harbor service.
// It contains a Client, which is used to lazy-load and cache data.
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

// Get returns the VM, client, and generator.
//
// Depending on the options specified, some return values may be nil.
// This is useful when you don't always need all the resources.
func (c *Client) Get(opts *opts.Opts) (*model.VM, *k8s.Client, generators.K8sGenerator, error) {
	var vm *model.VM
	var kc *k8s.Client
	var g generators.K8sGenerator
	var err error

	if opts.VmID != "" {
		vm, err = c.VM(opts.VmID, nil)
		if err != nil {
			return nil, nil, nil, err
		}

		if vm == nil {
			return nil, nil, nil, sErrors.VmNotFoundErr
		}
	}

	if opts.Client {
		zone := getZone(opts, vm)
		if zone == nil {
			return nil, nil, nil, sErrors.ZoneNotFoundErr
		}

		kc, err = c.Client(zone)
		if err != nil {
			return nil, nil, nil, err
		}

		if kc == nil {
			return nil, nil, nil, sErrors.VmNotFoundErr
		}
	}

	if opts.Generator {
		zone := getZone(opts, vm)
		if zone == nil {
			return nil, nil, nil, sErrors.ZoneNotFoundErr
		}

		g = c.Generator(vm, kc, zone, opts.ExtraOpts.ExtraSshKeys)
		if g == nil {
			return nil, nil, nil, sErrors.VmNotFoundErr
		}
	}

	return vm, kc, g, nil
}

// Client returns the K8s service client.
func (c *Client) Client(zone *configModels.Zone) (*k8s.Client, error) {
	return withClient(zone, getNamespaceName(zone))
}

// Generator returns the K8s generator.
func (c *Client) Generator(vm *model.VM, client *k8s.Client, zone *configModels.Zone, extraSshKeys []string) generators.K8sGenerator {
	if vm == nil {
		panic("vm is nil")
	}

	if client == nil {
		panic("client is nil")
	}

	if zone == nil {
		panic("zone is nil")
	}

	return resources.K8s(vm, zone, client, getNamespaceName(zone), extraSshKeys)
}

// getNamespaceName returns the namespace name
func getNamespaceName(zone *configModels.Zone) string {
	return zone.K8s.Namespaces.VM
}

// withClient returns a new K8s service client.
func withClient(zone *configModels.Zone, namespace string) (*k8s.Client, error) {
	k8sClient, err := k8s.New(&k8s.ClientConf{
		K8sClient:         zone.K8s.Client,
		KubeVirtK8sClient: zone.K8s.KubeVirtClient,
		Namespace:         namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client. details: %w", err)
	}

	return k8sClient, nil
}

// getZone is a helper function that returns either the zone in opts or the zone in vm.
func getZone(opts *opts.Opts, vm *model.VM) *configModels.Zone {
	if opts.ExtraOpts.Zone != nil {
		return opts.ExtraOpts.Zone
	}

	if vm != nil {
		return config.Config.GetZone(vm.Zone)
	}

	return nil
}
