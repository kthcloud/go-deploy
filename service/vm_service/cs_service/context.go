package cs_service

import (
	"fmt"
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
	makeError := func(err error) error {
		return fmt.Errorf("error creating cs service context. details: %w", err)
	}

	baseContext, err := base.NewVmBaseContext(vmID)
	if err != nil {
		return nil, makeError(err)
	}

	csClient, err := withCsClient(baseContext.Zone)
	if err != nil {
		return nil, makeError(err)
	}

	return &Context{
		VmContext: *baseContext,
		Client:    *csClient,
		Generator: *baseContext.Generator.CS(),
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
		Generator: *baseContext.Generator.CS(),
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
