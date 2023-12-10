package cs_service

import (
	configModels "go-deploy/models/config"
	"go-deploy/models/sys/vm"
	"go-deploy/pkg/config"
	"go-deploy/pkg/subsystems/cs"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/resources"
	"go-deploy/service/vm_service/client"
)

// Client is the client for the Harbor service.
// It contains a BaseClient, which is used to lazy-load and cache data.
type Client struct {
	client.BaseClient[Client]

	client    *cs.Client
	generator *resources.CsGenerator
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
func (c *Client) Get(opts *client.Opts) (*vm.VM, *cs.Client, *resources.CsGenerator, error) {
	var v *vm.VM
	var err error

	if opts.VM != "" {
		v, err = c.VM(opts.VM, nil)
		if err != nil {
			return nil, nil, nil, err
		}

		if v == nil {
			return nil, nil, nil, sErrors.VmNotFoundErr
		}
	}

	var cc *cs.Client
	if opts.Client {
		cc, err = c.GetOrCreateClient()
		if cc == nil {
			return nil, nil, nil, sErrors.VmNotFoundErr
		}
	}

	var g *resources.CsGenerator
	if opts.Generator {
		g = c.Generator(v)
		if g == nil {
			return nil, nil, nil, sErrors.VmNotFoundErr
		}
	}

	return v, cc, g, nil
}

// Client returns the GitHub service client.
//
// This does not create a new client if it does not exist.
func (c *Client) Client() *cs.Client {
	return c.client
}

// GetOrCreateClient returns the GitHub service client.
//
// If the client does not exist, it will be created.
func (c *Client) GetOrCreateClient() (*cs.Client, error) {
	if c.client == nil {
		if c.Zone == nil {
			panic("zone is nil")
		}

		csc, err := withCsClient(c.Zone)
		if err != nil {
			return nil, err
		}

		c.client = csc
	}

	return c.client, nil
}

// Generator returns the CS generator.
//
// If the generator does not exist, it will be created.
func (c *Client) Generator(vm *vm.VM) *resources.CsGenerator {
	if c.generator == nil {
		pg := resources.PublicGenerator()

		if vm != nil {
			pg.WithVM(vm)
		}

		if c.Zone != nil {
			pg.WithVmZone(c.Zone)
		}

		c.generator = pg.CS()
	}

	return c.generator
}

// withClient returns a new service client.
func withCsClient(zone *configModels.VmZone) (*cs.Client, error) {
	return cs.New(&cs.ClientConf{
		URL:         config.Config.CS.URL,
		ApiKey:      config.Config.CS.ApiKey,
		Secret:      config.Config.CS.Secret,
		IpAddressID: zone.IpAddressID,
		ProjectID:   zone.ProjectID,
		NetworkID:   zone.NetworkID,
		ZoneID:      zone.ZoneID,
	})
}
