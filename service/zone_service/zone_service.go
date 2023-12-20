package zone_service

import (
	configModels "go-deploy/models/config"
	zoneModel "go-deploy/models/sys/zone"
	"go-deploy/pkg/config"
	"go-deploy/service"
	"sort"
)

func (c *Client) List(opts ...ListOpts) ([]zoneModel.Zone, error) {
	o := service.GetFirstOrDefault(opts)

	deploymentZones := config.Config.Deployment.Zones
	vmZones := config.Config.VM.Zones

	var zones []zoneModel.Zone
	if o.Type != nil && *o.Type == zoneModel.ZoneTypeDeployment {
		zones = make([]zoneModel.Zone, len(deploymentZones))
		for i, zone := range deploymentZones {
			zones[i] = *toDZone(&zone)
		}
	} else if o.Type != nil && *o.Type == zoneModel.ZoneTypeVM {
		zones = make([]zoneModel.Zone, len(vmZones))
		for i, zone := range vmZones {
			zones[i] = *toVZone(&zone)
		}
	} else {
		zones = make([]zoneModel.Zone, len(deploymentZones)+len(vmZones))
		for i, zone := range deploymentZones {
			zones[i] = *toDZone(&zone)
		}

		for i, zone := range vmZones {
			zones[i+len(deploymentZones)] = *toVZone(&zone)
		}
	}

	sort.Slice(zones, func(i, j int) bool {
		return zones[i].Name < zones[j].Name || (zones[i].Name == zones[j].Name && zones[i].Type < zones[j].Type)
	})

	return zones, nil
}

func (c *Client) Get(name, zoneType string) *zoneModel.Zone {
	switch zoneType {
	case zoneModel.ZoneTypeDeployment:
		return toDZone(config.Config.Deployment.GetZone(name))
	case zoneModel.ZoneTypeVM:
		return toVZone(config.Config.VM.GetZone(name))
	}

	return nil
}

func toDZone(dZone *configModels.DeploymentZone) *zoneModel.Zone {
	if dZone == nil {
		return nil
	}

	domain := dZone.ParentDomain
	return &zoneModel.Zone{
		Name:        dZone.Name,
		Description: dZone.Description,
		Type:        zoneModel.ZoneTypeDeployment,
		Interface:   &domain,
	}
}

func toVZone(vmZone *configModels.VmZone) *zoneModel.Zone {
	if vmZone == nil {
		return nil
	}

	domain := vmZone.ParentDomain
	return &zoneModel.Zone{
		Name:        vmZone.Name,
		Description: vmZone.Description,
		Type:        zoneModel.ZoneTypeVM,
		Interface:   &domain,
	}
}
