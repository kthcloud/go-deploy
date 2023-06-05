package internal_service

import (
	"context"
	"encoding/json"
	"fmt"
	githubThirdParty "github.com/google/go-github/github"
	"github.com/google/uuid"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/github"
	githubModels "go-deploy/pkg/subsystems/github/models"
	"go-deploy/utils/requestutils"
	"log"
	"net/url"
	"strings"
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

func CreateGitHub(name string, params *deploymentModel.CreateParams) error {
	log.Println("setting up github for", name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup github for deployment %s. details: %s", name, err)
	}

	client, err := withGitHubClient(params.GitHub.Token)
	if err != nil {
		return makeError(err)
	}

	deployment, err := deploymentModel.GetByName(name)
	if err != nil {
		return makeError(err)
	}

	if deployment == nil {
		log.Println("deployment", name, "not found for github setup. assuming it was deleted")
		return nil
	}

	webhook := &deployment.Subsystems.GitHub.Webhook
	if !webhook.Created() {
		webhook, err = createGitHubWebhook(client, deployment, createGitHubWebhookPublic(params.GitHub.RepositoryID))
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

func DeleteGitHub(name string, githubToken *string) error {
	log.Println("deleting github for", name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to delete github for deployment %s. details: %s", name, err)
	}

	if githubToken == nil {
		// assume token is not attainable and that the webhook can remain active
		err := deploymentModel.UpdateSubsystemByName(name, "github", "placeholder", false)
		if err != nil {
			return makeError(err)
		}

		err = deploymentModel.UpdateSubsystemByName(name, "github", "webhook", githubModels.WebhookPublic{})
		if err != nil {
			return makeError(err)
		}
		return nil
	}

	client, err := github.New(&github.ClientConf{
		Token: *githubToken,
	})

	if err != nil {
		return makeError(err)
	}

	deployment, err := deploymentModel.GetByName(name)
	if err != nil {
		return makeError(err)
	}

	if deployment == nil {
		log.Println("deployment", name, "not found for github deletion. assuming it was deleted")
		return nil
	}

	if deployment.Subsystems.GitHub.Webhook.ID != 0 {
		err = client.DeleteWebhook(deployment.Subsystems.GitHub.Webhook.ID, deployment.Subsystems.GitHub.Webhook.RepositoryID)
		if err != nil {
			return makeError(err)
		}

		err = deploymentModel.UpdateSubsystemByName(name, "github", "webhook", githubModels.WebhookPublic{})
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

func CreatePlaceholderGitHub(name string) error {
	log.Println("setting up placeholder github")

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup placeholder github. details: %s", err)
	}

	err := deploymentModel.UpdateSubsystemByName(name, "github", "placeholder", true)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func ValidateGitHubToken(token string) (bool, string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to validate github token. details: %s", err)
	}

	client, err := withGitHubClient(token)
	if err != nil {
		return false, "", makeError(err)
	}

	limits, resp, err := client.GitHubClient.RateLimits(context.TODO())
	if err != nil {
		githubError := err.(*githubThirdParty.ErrorResponse)
		if githubError != nil && githubError.Message == "Bad credentials" {
			return false, "Invalid token", nil
		}

		return false, "", makeError(err)
	}

	if resp.StatusCode != 200 {
		return false, "", fmt.Errorf("failed to validate github token. status code: %d", resp.StatusCode)
	}

	if limits == nil {
		return false, "", fmt.Errorf("failed to validate github token. limits are nil")
	}

	if limits.Core.Remaining < 100 {
		return false, "Too few core requests remaining", nil
	}

	scopes := resp.Header.Get("X-OAuth-Scopes")
	if !strings.Contains(scopes, "admin:repo_hook") &&
		!strings.Contains(scopes, "repo") &&
		!strings.Contains(scopes, "user") {
		return false, "Requires scope to be one of: admin:repo_hook, repo, user", nil
	}

	return true, "", nil
}

func GetGitHubAccessTokenByCode(code string) (string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get github access token. details: %s", err)
	}

	apiRoute := "https://github.com/login/oauth/access_token"

	body := map[string]string{
		"client_id":     "6c3a489177c7833cc639",
		"client_secret": conf.Env.GitHub.ClientSecretProd,
		"code":          code,
	}

	bodyData, err := json.Marshal(body)
	if err != nil {
		return "", makeError(err)
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

func GetGitHubRepositories(token string) ([]deploymentModel.GitHubRepository, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get github repositories. details: %s", err)
	}

	client, err := withGitHubClient(token)
	if err != nil {
		return nil, makeError(err)
	}

	user, _, err := client.GitHubClient.Users.Get(context.Background(), "")
	if err != nil {
		return nil, makeError(err)
	}

	if user.Login == nil {
		return nil, makeError(fmt.Errorf("failed to get github repositories. user login is nil"))
	}

	repos, _, err := client.GitHubClient.Repositories.List(context.Background(), *user.Login, nil)
	if err != nil {
		return nil, makeError(err)
	}

	gitHubRepos := make([]deploymentModel.GitHubRepository, 0)
	for _, repo := range repos {
		if repo.ID != nil && repo.Name != nil {
			gitHubRepos = append(gitHubRepos, deploymentModel.GitHubRepository{
				ID:   *repo.ID,
				Name: *repo.Name,
			})
		}
	}

	return gitHubRepos, nil
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
