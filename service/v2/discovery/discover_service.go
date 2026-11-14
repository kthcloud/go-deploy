package discovery

import (
	"context"

	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/models/version"
	"github.com/kthcloud/go-deploy/pkg/config"
)

// Discover returns the discover information.
// This is stored locally in the config.
func (c *Client) Discover(_ context.Context) (*model.Discover, error) {
	return &model.Discover{
		Version: version.AppVersion,
		Roles:   config.Config.GetRoles(),
	}, nil
}
