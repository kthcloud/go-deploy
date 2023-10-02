package zone_service

import (
	zoneModel "go-deploy/models/sys/zone"
	"go-deploy/pkg/conf"
)

func GetAllZones() ([]zoneModel.Zone, error) {

	deploymentZones := conf.Env.Deployment.Zones
	vmZones := conf.Env.VM.Zones

	zones := make([]zoneModel.Zone, len(deploymentZones)+len(vmZones))

	for i, zone := range deploymentZones {
		domain := zone.ParentDomain
		zones[i] = zoneModel.Zone{
			Name:        zone.Name,
			Description: zone.Description,
			Type:        zoneModel.ZoneTypeDeployment,
			Interface:   &domain,
		}
	}

	for i, zone := range vmZones {
		domain := zone.ParentDomain
		zones[i+len(deploymentZones)] = zoneModel.Zone{
			Name:        zone.Name,
			Description: zone.Description,
			Type:        zoneModel.ZoneTypeVM,
			Interface:   &domain,
		}
	}

	return zones, nil
}

func GetZone(name, zoneType string) *zoneModel.Zone {
	zones, err := GetAllZones()
	if err != nil {
		return nil
	}

	for _, zone := range zones {
		if zone.Name == name && zone.Type == zoneType {
			return &zone
		}
	}

	return nil
}

func GetZonesByType(zoneType string) ([]zoneModel.Zone, error) {
	zones, err := GetAllZones()
	if err != nil {
		return nil, err
	}

	filteredZones := make([]zoneModel.Zone, 0)

	for _, zone := range zones {
		if zone.Type == zoneType {
			filteredZones = append(filteredZones, zone)
		}
	}

	return filteredZones, nil
}
