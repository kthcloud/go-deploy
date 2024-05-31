package discovery

import (
	"go-deploy/models/model"
	"go-deploy/models/version"
	"go-deploy/pkg/config"
)

// Discover returns the discover information.
// This is stored locally in the config.
func (c *Client) Discover() (*model.Discover, error) {
	return &model.Discover{
		Version: version.AppVersion,
		Roles:   config.Config.GetRoles(),
	}, nil
}
