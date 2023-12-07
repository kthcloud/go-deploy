package deployment_service

import (
	"go-deploy/pkg/config"
)

func ValidateHarborToken(secret string) bool {
	return secret == config.Config.Harbor.WebhookSecret
}
