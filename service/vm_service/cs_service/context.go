package cs_service

import (
	"go-deploy/models/sys/enviroment"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/cs"
	"go-deploy/service/resources"
	"go-deploy/service/vm_service/base"
)

type Context struct {
	base.VmContext

	Client    cs.Client
	Generator resources.CsGenerator
}

func NewContext(vmID string) (*Context, error) {
	baseContext, err := base.NewVmBaseContext(vmID)
	if err != nil {
		return nil, err
	}

	csClient, err := withCsClient(baseContext.Zone)
	if err != nil {
		return nil, err
	}

	return &Context{
		VmContext: *baseContext,
		Client:    *csClient,
		Generator: *baseContext.Generator.CS(func() (int, error) {
			return csClient.GetFreePort(baseContext.Zone.PortRange.Start, baseContext.Zone.PortRange.End)
		}),
	}, nil
}

func NewContextWithoutVM(zoneName string) (*Context, error) {
	baseContext, err := base.NewVmBaseContextWithoutVM(zoneName)
	if err != nil {
		return nil, err
	}

	csClient, err := withCsClient(baseContext.Zone)
	if err != nil {
		return nil, err
	}

	return &Context{
		VmContext: *baseContext,
		Client:    *csClient,
		Generator: *baseContext.Generator.CS(func() (int, error) {
			return csClient.GetFreePort(baseContext.Zone.PortRange.Start, baseContext.Zone.PortRange.End)
		}),
	}, nil

}

func withCsClient(zone *enviroment.VmZone) (*cs.Client, error) {
	return cs.New(&cs.ClientConf{
		URL:         conf.Env.CS.URL,
		ApiKey:      conf.Env.CS.ApiKey,
		Secret:      conf.Env.CS.Secret,
		IpAddressID: zone.IpAddressID,
		ProjectID:   zone.ProjectID,
		NetworkID:   zone.NetworkID,
		ZoneID:      zone.ZoneID,
	})
}
