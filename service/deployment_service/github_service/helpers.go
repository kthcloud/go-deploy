package github_service

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/github"
	githubModels "go-deploy/pkg/subsystems/github/models"
	"go-deploy/utils/requestutils"
	"net/url"
)

func createGitHubWebhookPublic(repositoryID int64) *githubModels.WebhookPublic {
	webhookTarget := fmt.Sprintf("%s/v1/hooks/deployments/github", conf.Env.ExternalUrl)
	return &githubModels.WebhookPublic{
		RepositoryID: repositoryID,
		Events:       nil,
		Active:       false,
		ContentType:  "json",
		WebhookURL:   webhookTarget,
		Secret:       uuid.NewString(),
	}
}

func withGitHubClient(token string) (*github.Client, error) {
	return github.New(&github.ClientConf{
		Token: token,
	})
}

func createGitHubWebhook(client *github.Client, deployment *deploymentModel.Deployment, public *githubModels.WebhookPublic) (*githubModels.WebhookPublic, error) {
	id, err := client.CreateWebhook(public)
	if err != nil {
		return nil, err
	}

	webhook, err := client.ReadWebhook(id, public.RepositoryID)
	if err != nil {
		return nil, err
	}

	if webhook == nil {
		return nil, fmt.Errorf("failed to read webhook after creation")
	}

	err = deploymentModel.UpdateSubsystemByName(deployment.Name, "github", "webhook", webhook)
	if err != nil {
		return nil, err
	}

	return webhook, nil
}

func fetchAccessToken(code, clientId string, clientSecret string) (string, error) {
	apiRoute := "https://github.com/login/oauth/access_token"

	body := map[string]string{
		"client_id":     clientId,
		"client_secret": clientSecret,
		"code":          code,
	}

	bodyData, err := json.Marshal(body)
	if err != nil {
		return "", err
	}

	res, err := requestutils.DoRequest("POST", apiRoute, bodyData, nil)
	if err != nil {
		return "", err
	}

	if res.StatusCode != 200 {
		return "", fmt.Errorf("failed to get github access token. status code: %d", res.StatusCode)
	}

	readBody, err := requestutils.ReadBody(res.Body)
	if err != nil {
		return "", err
	}

	paramsStrings := string(readBody)

	params, err := url.ParseQuery(paramsStrings)
	if err != nil {
		return "", err
	}

	accessToken := params.Get("access_token")
	if accessToken == "" {
		return "", fmt.Errorf("failed to get github access token. access token is empty")
	}

	return accessToken, nil
}
