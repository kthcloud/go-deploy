package zones

import (
	configModels "go-deploy/models/config"
	"go-deploy/pkg/config"
	"go-deploy/service/utils"
	"go-deploy/service/v1/zones/opts"
)

// Get gets a zone by name and type
func (c *Client) Get(name string) *configModels.Zone {
	return config.Config.GetZone(name)
}

// GetLegacy gets a legacy zone by name
func (c *Client) GetLegacy(name string) *configModels.LegacyZone {
	return config.Config.VM.GetLegacyZone(name)
}

// List gets a list of zones
func (c *Client) List(opts ...opts.ListOpts) ([]configModels.Zone, error) {
	_ = utils.GetFirstOrDefault(opts)

	return config.Config.Zones, nil
}

// ListLegacy gets a list of legacy zones
func (c *Client) ListLegacy(opts ...opts.ListOpts) ([]configModels.LegacyZone, error) {
	_ = utils.GetFirstOrDefault(opts)

	return config.Config.VM.Zones, nil
}

func (c *Client) HasCapability(zoneName, capability string) bool {
	zone := config.Config.GetZone(zoneName)
	if zone == nil {
		return false
	}

	return zone.HasCapability(capability)
}
