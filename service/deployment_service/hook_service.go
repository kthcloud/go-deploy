package deployment_service

import (
	"go-deploy/models/dto/body"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/pkg/conf"
)

func GetAllByGitHubWebhookID(id int64) ([]deploymentModel.Deployment, error) {
	return deploymentModel.GetAllByGitHubWebhookID(id)
}

func ValidateHarborToken(secret string) bool {
	return secret == conf.Env.Harbor.WebhookSecret
}

func GetByHarborWebhook(webhook *body.HarborWebhook) (*deploymentModel.Deployment, error) {
	return deploymentModel.GetByName(webhook.EventData.Repository.Name)
}
