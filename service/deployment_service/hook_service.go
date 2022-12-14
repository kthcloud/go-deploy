package deployment_service

import (
	"go-deploy/models/deployment"
	"go-deploy/models/dto"
	"go-deploy/utils/requestutils"
	"io"
)

func GetHook(body io.ReadCloser) (*dto.HarborWebhook, error) {
	readBody, err := requestutils.ReadBody(body)
	if err != nil {
		return nil, err
	}

	var webhook = dto.HarborWebhook{}
	err = requestutils.ParseJson(readBody, &webhook)
	if err != nil {
		return nil, err
	}

	return &webhook, nil
}

func GetByWebhookToken(token string) (*deployment.Deployment, error) {
	return deployment.GetByWebhookToken(token)
}
