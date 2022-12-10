package deployment_service

import (
	"go-deploy/models"
	"go-deploy/models/dto"
	"go-deploy/utils/requestutils"
	"io"
)

func GetHook(body io.ReadCloser) (*dto.HarborWebook, error) {
	readBody, err := requestutils.ReadBody(body)
	if err != nil {
		return nil, err
	}

	var webhook = dto.HarborWebook{}
	err = requestutils.ParseJson(readBody, &webhook)
	if err != nil {
		return nil, err
	}

	return &webhook, nil
}

func GetByWebhookToken(token string) (*models.Deployment, error) {
	return models.GetByWebookToken(token)
}
