package zones

import (
	configModels "go-deploy/models/config"
	zoneModels "go-deploy/models/sys/zone"
	"go-deploy/pkg/config"
	"go-deploy/service/utils"
	"go-deploy/service/v1/zones/opts"
	"sort"
)

// Get gets a zone by name and type
func (c *Client) Get(name, zoneType string) *zoneModels.Zone {
	switch zoneType {
	case zoneModels.TypeDeployment:
		return toDZone(config.Config.Deployment.GetZone(name))
	case zoneModels.TypeVM:
		return toVZone(config.Config.VM.GetZone(name))
	}

	return nil
}

// List gets a list of zones
func (c *Client) List(opts ...opts.ListOpts) ([]zoneModels.Zone, error) {
	o := utils.GetFirstOrDefault(opts)

	deploymentZones := config.Config.Deployment.Zones
	vmZones := config.Config.VM.Zones

	var zones []zoneModels.Zone
	if o.Type != nil && *o.Type == zoneModels.TypeDeployment {
		zones = make([]zoneModels.Zone, len(deploymentZones))
		for i, zone := range deploymentZones {
			zones[i] = *toDZone(&zone)
		}
	} else if o.Type != nil && *o.Type == zoneModels.TypeVM {
		zones = make([]zoneModels.Zone, len(vmZones))
		for i, zone := range vmZones {
			zones[i] = *toVZone(&zone)
		}
	} else {
		zones = make([]zoneModels.Zone, len(deploymentZones)+len(vmZones))
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

func toDZone(dZone *configModels.DeploymentZone) *zoneModels.Zone {
	if dZone == nil {
		return nil
	}

	domain := dZone.ParentDomain
	return &zoneModels.Zone{
		Name:        dZone.Name,
		Description: dZone.Description,
		Type:        zoneModels.TypeDeployment,
		Interface:   &domain,
	}
}

func toVZone(vmZone *configModels.VmZone) *zoneModels.Zone {
	if vmZone == nil {
		return nil
	}

	domain := vmZone.ParentDomain
	return &zoneModels.Zone{
		Name:        vmZone.Name,
		Description: vmZone.Description,
		Type:        zoneModels.TypeVM,
		Interface:   &domain,
	}
}