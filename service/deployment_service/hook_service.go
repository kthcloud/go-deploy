package deployment_service

import (
	"go-deploy/models/deployment"
	bodyDto "go-deploy/models/dto/body"
	"go-deploy/utils/requestutils"
	"io"
	"log"
)

func GetHook(body io.ReadCloser) (*bodyDto.HarborWebhook, error) {
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

func GetByWebhookToken(token string) (*deployment.Deployment, error) {
	return deployment.GetByWebhookToken(token)
}
