package deployment_service

import (
	"go-deploy/pkg/config"
)

// ValidateHarborToken validates the Harbor token.
func ValidateHarborToken(secret string) bool {
	return secret == config.Config.Harbor.WebhookSecret
}
