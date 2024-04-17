package cs_service

import (
	configModels "go-deploy/models/config"
	"go-deploy/models/model"
	"go-deploy/pkg/config"
	"go-deploy/pkg/subsystems/cs"
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

// OptsOnlyClient returns the options required to get only the client.
func OptsOnlyClient(zone *configModels.LegacyZone) *opts.Opts {
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
func (c *Client) Get(opts *opts.Opts) (*model.VM, *cs.Client, generators.CsGenerator, error) {
	var vm *model.VM
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

	var csClient *cs.Client
	if opts.Client {
		zone := getZone(opts, vm)
		if zone == nil {
			return nil, nil, nil, sErrors.ZoneNotFoundErr
		}

		csClient, err = c.Client(zone)
		if csClient == nil {
			return nil, nil, nil, sErrors.VmNotFoundErr
		}
	}

	var csGenerator generators.CsGenerator
	if opts.Generator {
		zone := getZone(opts, vm)
		if zone == nil {
			return nil, nil, nil, sErrors.ZoneNotFoundErr
		}

		csGenerator = c.Generator(vm, zone)
		if csGenerator == nil {
			return nil, nil, nil, sErrors.VmNotFoundErr
		}
	}

	return vm, csClient, csGenerator, nil
}

// Client returns the CloudStack service client.
//
// If WithZone is set, it will try to use those values.
// Otherwise, it will use the zone parameter.
// If both are nil, it will panic.
func (c *Client) Client(zone *configModels.LegacyZone) (*cs.Client, error) {
	if zone == nil {
		panic("zone is nil")
	}

	return withCsClient(zone)
}

// Generator returns the CS generator.
func (c *Client) Generator(vm *model.VM, zone *configModels.LegacyZone) generators.CsGenerator {
	if vm == nil {
		panic("vm is nil")
	}

	if zone == nil {
		panic("zone is nil")
	}

	return resources.CS(vm, zone)
}

// withClient returns a new service client.
func withCsClient(zone *configModels.LegacyZone) (*cs.Client, error) {
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

// getZone is a helper function that returns either the zone in opts or the zone in vm.
func getZone(opts *opts.Opts, vm *model.VM) *configModels.LegacyZone {
	if opts.ExtraOpts.Zone != nil {
		return opts.ExtraOpts.Zone
	}

	if vm != nil {
		return config.Config.VM.GetLegacyZone(vm.Zone)
	}

	return nil
}
