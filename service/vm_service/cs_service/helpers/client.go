package helpers

import (
	"fmt"
	"go-deploy/models/sys/enviroment"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/models/sys/vm/subsystems"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/cs"
)

type Client struct {
	UpdateDB func(string, string, interface{}) error
	// subsystem client
	SsClient *cs.Client
	Zone     *enviroment.VmZone
	CS       *subsystems.CS
}

func New(cs *subsystems.CS, zoneName string) (*Client, error) {
	makeError := func(err error) error {
		return fmt.Errorf("error creating cs client in vm helper client. details: %w", err)
	}

	zone := conf.Env.VM.GetZone(zoneName)
	if zone == nil {
		return nil, makeError(fmt.Errorf("zone %s not found", zoneName))
	}

	csClient, err := withCsClient(zone)
	if err != nil {
		return nil, makeError(err)
	}

	return &Client{
		UpdateDB: func(id, key string, data interface{}) error {
			return vmModel.New().UpdateSubsystemByID(id, "cs", key, data)
		},
		SsClient: csClient,
		Zone:     zone,
		CS:       cs,
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
