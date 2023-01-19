package deployment_service

import (
	"go-deploy/models/deployment"
	"go-deploy/models/dto"
	"go-deploy/utils/requestutils"
	"io"
	"log"
)

func GetHook(body io.ReadCloser) (*dto.HarborWebhook, error) {
	readBody, err := requestutils.ReadBody(body)
	if err != nil {
		return nil, err
	}

	log.Println(string(readBody))

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
