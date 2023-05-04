package deployment_service

import (
	bodyDto "go-deploy/models/dto/body"
	deployment2 "go-deploy/models/sys/deployment"
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

func GetByWebhookToken(token string) (*deployment2.Deployment, error) {
	return deployment2.GetByWebhookToken(token)
}
