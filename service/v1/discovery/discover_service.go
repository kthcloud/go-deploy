package discovery

import (
	"go-deploy/models/model"
	"go-deploy/pkg/config"
)

// Discover returns the discover information.
// This is stored locally in the config.
func (c *Client) Discover() (*model.Discover, error) {
	return &model.Discover{
		Version: config.Config.Version,
		Roles:   config.Config.Roles,
	}, nil
}
