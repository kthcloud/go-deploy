package zones

import (
	configModels "go-deploy/models/config"
	"go-deploy/models/model"
	"go-deploy/pkg/config"
	"go-deploy/service/utils"
	"go-deploy/service/v1/zones/opts"
	"sort"
)

// Get gets a zone by name and type
func (c *Client) Get(name, zoneType string) *model.Zone {
	switch zoneType {
	case model.ZoneTypeDeployment:
		return toDZone(config.Config.GetZone(name))
	case model.ZoneTypeVM:
		return toVZone(config.Config.VM.GetLegacyZone(name))
	}

	return nil
}

// List gets a list of zones
func (c *Client) List(opts ...opts.ListOpts) ([]model.Zone, error) {
	o := utils.GetFirstOrDefault(opts)

	deploymentZones := config.Config.Zones
	vmZones := config.Config.VM.Zones

	var zones []model.Zone
	if o.Type != nil && *o.Type == model.ZoneTypeDeployment {
		zones = make([]model.Zone, len(deploymentZones))
		for i, zone := range deploymentZones {
			zones[i] = *toDZone(&zone)
		}
	} else if o.Type != nil && *o.Type == model.ZoneTypeVM {
		zones = make([]model.Zone, len(vmZones))
		for i, zone := range vmZones {
			zones[i] = *toVZone(&zone)
		}
	} else {
		zones = make([]model.Zone, len(deploymentZones)+len(vmZones))
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

func (c *Client) HasCapability(zoneName, capability string) bool {
	zone := config.Config.GetZone(zoneName)
	if zone == nil {
		return false
	}

	return zone.HasCapability(capability)
}

func toDZone(dZone *configModels.Zone) *model.Zone {
	if dZone == nil {
		return nil
	}

	domain := dZone.Domains.ParentDeployment
	return &model.Zone{
		Name:        dZone.Name,
		Description: dZone.Description,
		Interface:   &domain,
		Type:        model.ZoneTypeDeployment,
	}
}

func toVZone(vmZone *configModels.LegacyZone) *model.Zone {
	if vmZone == nil {
		return nil
	}

	domain := vmZone.ParentDomain
	return &model.Zone{
		Name:        vmZone.Name,
		Description: vmZone.Description,
		Interface:   &domain,
		Type:        model.ZoneTypeVM,
	}
}
