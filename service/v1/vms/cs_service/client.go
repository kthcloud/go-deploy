package cs_service

import (
	configModels "go-deploy/models/config"
	vmModels "go-deploy/models/sys/vm"
	"go-deploy/pkg/config"
	"go-deploy/pkg/subsystems/cs"
	"go-deploy/service/core"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/resources"
	"go-deploy/service/v1/vms/client"
	"go-deploy/service/v1/vms/opts"
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

// OptsOnlyClient returns the options required to get only the client.
func OptsOnlyClient(zone *configModels.VmZone) *opts.Opts {
	return &opts.Opts{
		Client: true,
		ExtraOpts: opts.ExtraOpts{
			Zone: zone,
		},
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

// Get is a helper function returns resources that assists with interacting with the subsystem.
// Essentially just collector the VM, client and generator.
//
// Depending on the options specified, some return values may be nil.
// This is useful when you don't always need all the resources.
func (c *Client) Get(opts *opts.Opts) (*vmModels.VM, *cs.Client, *resources.CsGenerator, error) {
	var vm *vmModels.VM
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

	var cc *cs.Client
	if opts.Client {
		var zone *configModels.VmZone
		if opts.ExtraOpts.Zone != nil {
			zone = opts.ExtraOpts.Zone
		} else if vm != nil {
			zone = config.Config.VM.GetZone(vm.Zone)
		}

		cc, err = c.Client(zone)
		if cc == nil {
			return nil, nil, nil, sErrors.VmNotFoundErr
		}
	}

	var g *resources.CsGenerator
	if opts.Generator {
		var zone *configModels.VmZone
		if opts.ExtraOpts.Zone != nil {
			zone = opts.ExtraOpts.Zone
		} else if vm != nil {
			zone = config.Config.VM.GetZone(vm.Zone)
		}

		g = c.Generator(vm, zone)
		if g == nil {
			return nil, nil, nil, sErrors.VmNotFoundErr
		}
	}

	return vm, cc, g, nil
}

// Client returns the CloudStack service client.
//
// If WithZone is set, it will try to use those values.
// Otherwise, it will use the zone parameter.
// If both are nil, it will panic.
func (c *Client) Client(zone *configModels.VmZone) (*cs.Client, error) {
	if zone == nil {
		panic("zone is nil")
	}

	return withCsClient(zone)
}

// Generator returns the CS generator.
func (c *Client) Generator(vm *vmModels.VM, zone *configModels.VmZone) *resources.CsGenerator {
	if vm == nil {
		panic("vm is nil")
	}

	if zone == nil {
		panic("zone is nil")
	}

	return resources.PublicGenerator().WithVM(vm).WithVmZone(zone).CS()
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
		TemplateID:  zone.TemplateID,
	})
}
