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
		return toDZone(config.Config.Deployment.GetZone(name))
	case model.ZoneTypeVM:
		return toVZone(config.Config.VM.GetZone(name))
	}

	return nil
}

// List gets a list of zones
func (c *Client) List(opts ...opts.ListOpts) ([]model.Zone, error) {
	o := utils.GetFirstOrDefault(opts)

	deploymentZones := config.Config.Deployment.Zones
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

func toDZone(dZone *configModels.DeploymentZone) *model.Zone {
	if dZone == nil {
		return nil
	}

	domain := dZone.ParentDomain
	return &model.Zone{
		Name:        dZone.Name,
		Description: dZone.Description,
		Type:        model.ZoneTypeDeployment,
		Interface:   &domain,
	}
}

func toVZone(vmZone *configModels.VmZone) *model.Zone {
	if vmZone == nil {
		return nil
	}

	domain := vmZone.ParentDomain
	return &model.Zone{
		Name:        vmZone.Name,
		Description: vmZone.Description,
		Type:        model.ZoneTypeVM,
		Interface:   &domain,
	}
}
