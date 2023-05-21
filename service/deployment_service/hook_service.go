package deployment_service

import (
	bodyDto "go-deploy/models/dto/body"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/utils/requestutils"
	"io"
	"log"
)

func ParseHarborWebhook(body io.ReadCloser) (*bodyDto.HarborWebhook, error) {
	readBody, err := requestutils.ReadBody(body)
	if err != nil {
		return nil, err
	}

	log.Println(string(readBody))

	var webhook = bodyDto.HarborWebhook{}
	err = requestutils.ParseJson(readBody, &webhook)
	if err != nil {
		return nil, err
	}

	return &webhook, nil
}

func GetByHarborWebhookToken(token string) (*deploymentModel.Deployment, error) {
	return deploymentModel.GetByHarborWebhookToken(token)
}

func GetByGitHubWebhookID(id int64) (*deploymentModel.Deployment, error) {
	return deploymentModel.GetByGitHubWebhookID(id)
}
