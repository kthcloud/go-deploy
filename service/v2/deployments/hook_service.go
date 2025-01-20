package deployments

import (
	"github.com/kthcloud/go-deploy/pkg/config"
)

// ValidateHarborToken validates the Harbor token.
func (c *Client) ValidateHarborToken(secret string) bool {
	return secret == config.Config.Harbor.WebhookSecret
}
