package system

import (
	configModels "github.com/kthcloud/go-deploy/models/config"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/service/utils"
	"github.com/kthcloud/go-deploy/service/v2/system/opts"
)

// GetZone gets a zone by name and type
func (c *Client) GetZone(name string) *configModels.Zone {
	return config.Config.GetZone(name)
}

// ListZones gets a list of zones
func (c *Client) ListZones(opts ...opts.ListOpts) ([]configModels.Zone, error) {
	_ = utils.GetFirstOrDefault(opts)

	return config.Config.Zones, nil
}

func (c *Client) ZoneHasCapability(zoneName, capability string) bool {
	zone := config.Config.GetZone(zoneName)
	if zone == nil {
		return false
	}

	return zone.HasCapability(capability)
}
