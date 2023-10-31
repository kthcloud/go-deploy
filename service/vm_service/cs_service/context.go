package cs_service

import (
	"fmt"
	configModels "go-deploy/models/config"
	"go-deploy/pkg/config"
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
