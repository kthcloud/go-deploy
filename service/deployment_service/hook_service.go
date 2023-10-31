package deployment_service

import (
	"go-deploy/models/dto/body"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/pkg/config"
)

func GetAllByGitHubWebhookID(id int64) ([]deploymentModel.Deployment, error) {
	return deploymentModel.New().GetAllByGitHubWebhookID(id)
}

func ValidateHarborToken(secret string) bool {
	return secret == config.Config.Harbor.WebhookSecret
}

func GetByHarborWebhook(webhook *body.HarborWebhook) (*deploymentModel.Deployment, error) {
	return deploymentModel.New().GetByName(webhook.EventData.Repository.Name)
}
