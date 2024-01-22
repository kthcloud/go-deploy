package k8s_service

import (
	"fmt"
	configModels "go-deploy/models/config"
	vmModels "go-deploy/models/sys/vm"
	"go-deploy/pkg/config"
	"go-deploy/pkg/subsystems/k8s"
	"go-deploy/service"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/resources"
	"go-deploy/service/vm_service/client"
	"go-deploy/utils/subsystemutils"
)

// OptsAll returns the options required to get all the service tools, ie. VM, client and generator.
func OptsAll(vmID string, overwriteOps ...client.ExtraOpts) *client.Opts {
	var ow client.ExtraOpts
	if len(overwriteOps) > 0 {
		ow = overwriteOps[0]
	}

	return &client.Opts{
		VmID:      vmID,
		Client:    true,
		Generator: true,
		ExtraOpts: ow,
	}
}

// OptsNoGenerator returns the options required to get only the VM and client.
func OptsNoGenerator(vmID string, extraOpts ...client.ExtraOpts) *client.Opts {
	var eo client.ExtraOpts
	if len(extraOpts) > 0 {
		eo = extraOpts[0]
	}

	return &client.Opts{
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
func New(cache *service.Cache) *Client {
	c := &Client{BaseClient: client.NewBaseClient[Client](cache)}
	c.BaseClient.SetParent(c)
	return c
}

// Get returns the deployment, client, and generator.
//
// Depending on the options specified, some return values may be nil.
// This is useful when you don't always need all the resources.
func (c *Client) Get(opts *client.Opts) (*vmModels.VM, *k8s.Client, *resources.K8sGenerator, error) {
	var vm *vmModels.VM
	var kc *k8s.Client
	var g *resources.K8sGenerator
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

		var zone *configModels.DeploymentZone
		if opts.ExtraOpts.Zone != nil {
			zone = opts.ExtraOpts.DeploymentZone
		} else if vm != nil && vm.DeploymentZone != nil {
			zone = config.Config.Deployment.GetZone(*vm.DeploymentZone)
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
		var zone *configModels.VmZone
		if opts.ExtraOpts.Zone != nil {
			zone = opts.ExtraOpts.Zone
		} else if vm != nil {
			zone = config.Config.VM.GetZone(vm.Zone)
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

	return withClient(zone, getNamespaceName(userID))
}

// Generator returns the K8s generator.
func (c *Client) Generator(vm *vmModels.VM, client *k8s.Client, zone *configModels.VmZone, deploymentZone *configModels.DeploymentZone) *resources.K8sGenerator {
	if vm == nil {
		panic("vm is nil")
	}

	if client == nil {
		panic("client is nil")
	}

	if zone == nil {
		panic("zone is nil")
	}

	return resources.PublicGenerator().WithVM(vm).WithVmZone(zone).WithDeploymentZone(deploymentZone).K8s(client)
}

// getNamespaceName returns the namespace name for the user.
func getNamespaceName(userID string) string {
	return subsystemutils.GetPrefixedName(fmt.Sprintf("vm-%s", userID))
}

// withClient returns a new K8s service client.
func withClient(zone *configModels.DeploymentZone, namespace string) (*k8s.Client, error) {
	k8sClient, err := k8s.New(&k8s.ClientConf{
		K8sClient:     zone.Client,
		VirtK8sClient: nil,
		Namespace:     namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client. details: %w", err)
	}

	return k8sClient, nil
}
