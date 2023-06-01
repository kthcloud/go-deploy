package internal_service

import (
	"context"
	"fmt"
	githubThirdParty "github.com/google/go-github/github"
	"github.com/google/uuid"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/github"
	githubModels "go-deploy/pkg/subsystems/github/models"

	"log"
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
	if !strings.Contains(scopes, "admin:repo_hook") {
		return false, "Missing admin:repo_hook scope", nil
	}

	return true, "", nil
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
