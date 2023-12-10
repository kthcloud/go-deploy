package k8s_service

import (
	"fmt"
	configModels "go-deploy/models/config"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/pkg/subsystems/k8s"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/resources"
	"go-deploy/service/vm_service/client"
	"go-deploy/utils/subsystemutils"
)

// Client is the client for the Harbor service.
// It contains a BaseClient, which is used to lazy-load and cache data.
type Client struct {
	client.BaseClient[Client]

	client    *k8s.Client
	generator *resources.K8sGenerator
}

// New creates a new Client.
// If context is not nil, it will be used to create a new BaseClient.
// Otherwise, an empty context will be created.
func New(context *client.Context) *Client {
	c := &Client{
		BaseClient: client.NewBaseClient[Client](context),
	}
	c.BaseClient.SetParent(c)
	return c
}

// Get is a helper function returns resources that assists with interacting with the subsystem.
// Essentially just collector the VM, client and generator.
//
// Depending on the options specified, some return values may be nil.
// This is useful when you don't always need all the resources.
func (c *Client) Get(opts *client.Opts) (*vmModel.VM, *k8s.Client, *resources.K8sGenerator, error) {
	var vm *vmModel.VM
	var err error

	if opts.VM != "" {
		vm, err = c.VM(opts.VM, nil)
		if err != nil {
			return nil, nil, nil, err
		}

		if vm == nil {
			return nil, nil, nil, sErrors.VmNotFoundErr
		}
	}

	var kc *k8s.Client
	if opts.Client {
		kc, err = c.GetOrCreateClient()
		if kc == nil {
			return nil, nil, nil, sErrors.VmNotFoundErr
		}
	}

	var g *resources.K8sGenerator
	if opts.Generator {
		g = c.Generator(vm)
		if g == nil {
			return nil, nil, nil, sErrors.VmNotFoundErr
		}
	}

	return vm, kc, g, nil
}

// WithUserID sets the user id
// This is used to set the namespace
func (c *Client) WithUserID(userID string) *Client {
	if c.client != nil {
		c.client.Namespace = getNamespaceName(userID)
	}

	if c.generator != nil {
		c.generator.K8s(c.Client())
	}

	c.UserID = userID

	return c
}

// Client returns the K8s service client.
//
// This does not create a new client if it does not exist.
func (c *Client) Client() *k8s.Client {
	return c.client
}

// GetOrCreateClient returns the K8s service client.
//
// If the client does not exist, it will be created.
func (c *Client) GetOrCreateClient() (*k8s.Client, error) {
	if c.client == nil {
		if c.UserID == "" {
			panic("user id is empty")
		}

		if c.DeploymentZone == nil {
			panic("deployment zone is nil")
		}

		kc, err := withClient(c.DeploymentZone, getNamespaceName(c.UserID))
		if err != nil {
			return nil, err
		}

		c.client = kc
	}

	return c.client, nil
}

// Generator returns the K8s generator.
//
// If the generator does not exist, it will be created.
// If creating a new generator, the current deployment and zone will be used.
// Set the deployment and zone before calling this function by using WithDeployment and WithZone.
func (c *Client) Generator(vm *vmModel.VM) *resources.K8sGenerator {
	if c.generator == nil {
		pg := resources.PublicGenerator()

		if vm != nil {
			pg.WithVM(vm)
		}

		if c.Zone != nil {
			pg.WithVmZone(c.Zone)
		}

		if c.DeploymentZone != nil {
			pg.WithDeploymentZone(c.DeploymentZone)
		}

		c.generator = pg.K8s(c.Client())
	}

	return c.generator
}

// getNamespaceName returns the namespace name for the user.
func getNamespaceName(userID string) string {
	return subsystemutils.GetPrefixedName(userID)
}

// withClient returns a new K8s service client.
func withClient(zone *configModels.DeploymentZone, namespace string) (*k8s.Client, error) {
	k8sClient, err := k8s.New(zone.Client, namespace)
	if err != nil {
		return nil, fmt.Errorf("failed to create k8s client. details: %w", err)
	}

	return k8sClient, nil
}
