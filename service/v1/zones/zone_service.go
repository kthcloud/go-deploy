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
func (c *Client) Get(name string) *model.Zone {
	return toZone(config.Config.GetZone(name))
}

// GetLegacy gets a legacy zone by name
func (c *Client) GetLegacy(name string) *model.Zone {
	return toLegacyZone(config.Config.VM.GetLegacyZone(name))
}

// List gets a list of zones
func (c *Client) List(opts ...opts.ListOpts) ([]model.Zone, error) {
	_ = utils.GetFirstOrDefault(opts)

	deploymentZones := config.Config.Zones
	vmZones := config.Config.VM.Zones

	var zones []model.Zone
	zones = make([]model.Zone, len(deploymentZones)+len(vmZones))
	for i, zone := range deploymentZones {
		zones[i] = *toZone(&zone)
	}

	for i, zone := range vmZones {
		zones[i+len(deploymentZones)] = *toLegacyZone(&zone)
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

func toZone(zone *configModels.Zone) *model.Zone {
	if zone == nil {
		return nil
	}

	domain := zone.Domains.ParentDeployment
	return &model.Zone{
		Name:         zone.Name,
		Description:  zone.Description,
		Capabilities: zone.Capabilities,
		Interface:    &domain,
		Legacy:       false,

		Type: model.ZoneTypeDeployment,
	}
}

func toLegacyZone(legacyZone *configModels.LegacyZone) *model.Zone {
	if legacyZone == nil {
		return nil
	}

	domain := legacyZone.ParentDomain
	return &model.Zone{
		Name:        legacyZone.Name,
		Description: legacyZone.Description,
		Capabilities: []string{
			configModels.ZoneCapabilityVM,
		},
		Interface: &domain,
		Legacy:    true,

		Type: model.ZoneTypeVM,
	}
}
