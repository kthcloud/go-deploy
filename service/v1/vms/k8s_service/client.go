package k8s_service

import (
	"fmt"
	configModels "go-deploy/models/config"
	"go-deploy/models/model"
	"go-deploy/pkg/config"
	"go-deploy/pkg/subsystems/k8s"
	"go-deploy/service/core"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/generators"
	"go-deploy/service/v1/vms/client"
	"go-deploy/service/v1/vms/opts"
	"go-deploy/service/v1/vms/resources"
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

// Client is the client for the Harbor service.
// It contains a Client, which is used to lazy-load and cache data.
type Client struct {
	client.BaseClient[Client]
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
			return nil, nil, nil, sErrors.DeploymentNotFoundErr
		}
	}

	if opts.Client {
		var userID string
		if opts.ExtraOpts.UserID != "" {
			userID = opts.ExtraOpts.UserID
		} else {
			userID = vm.OwnerID
		}

		zone := getDeploymentZone(opts, vm)
		if zone == nil {
			return nil, nil, nil, sErrors.ZoneNotFoundErr
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
		zone := getVmZone(opts, vm)
		if zone == nil {
			return nil, nil, nil, sErrors.ZoneNotFoundErr
		}

		var deploymentZone *configModels.DeploymentZone
		if opts.ExtraOpts.DeploymentZone != nil {
			deploymentZone = opts.ExtraOpts.DeploymentZone
		} else if vm != nil && vm.DeploymentZone != nil {
			deploymentZone = config.Config.Deployment.GetZone(*vm.DeploymentZone)
		}

		g = c.Generator(vm, kc, zone, deploymentZone)
		if g == nil {
			return nil, nil, nil, sErrors.DeploymentNotFoundErr
		}
	}

	return vm, kc, g, nil
}

// Client returns the K8s service client.
func (c *Client) Client(userID string, zone *configModels.DeploymentZone) (*k8s.Client, error) {
	if userID == "" {
		panic("user id is empty")
	}

	return withClient(zone, getNamespaceName(zone))
}

// Generator returns the K8s generator.
func (c *Client) Generator(vm *model.VM, client *k8s.Client, zone *configModels.VmZone, deploymentZone *configModels.DeploymentZone) generators.K8sGenerator {
	if vm == nil {
		panic("vm is nil")
	}

	if client == nil {
		panic("client is nil")
	}

	if zone == nil {
		panic("zone is nil")
	}

	return resources.K8s(vm, zone, deploymentZone, client, getNamespaceName(deploymentZone))
}

// getNamespaceName returns the namespace name.
func getNamespaceName(zone *configModels.DeploymentZone) string {
	return zone.Namespaces.VM
}

// withClient returns a new K8s service client.
func withClient(zone *configModels.DeploymentZone, namespace string) (*k8s.Client, error) {
	k8sClient, err := k8s.New(&k8s.ClientConf{
		K8sClient:         zone.K8sClient,
		KubeVirtK8sClient: zone.KubeVirtClient,
		Namespace:         namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client. details: %w", err)
	}

	return k8sClient, nil
}

// getVmZone is a helper function that returns either the zone in opts or the zone in vm.
func getVmZone(opts *opts.Opts, vm *model.VM) *configModels.VmZone {
	if opts.ExtraOpts.Zone != nil {
		return opts.ExtraOpts.Zone
	}

	if vm != nil {
		return config.Config.VM.GetZone(vm.Zone)
	}

	return nil
}

// getDeploymentZone is a helper function that returns either the zone in opts or the zone in vm.
func getDeploymentZone(opts *opts.Opts, vm *model.VM) *configModels.DeploymentZone {
	if opts.ExtraOpts.DeploymentZone != nil {
		return opts.ExtraOpts.DeploymentZone
	}

	if vm != nil && vm.DeploymentZone != nil {
		return config.Config.Deployment.GetZone(*vm.DeploymentZone)
	}

	return nil
}
