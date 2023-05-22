package internal_service

import (
	"fmt"
	"github.com/google/uuid"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/github"
	githubModels "go-deploy/pkg/subsystems/github/models"
	"log"
)

func CreateGitHub(name string, params *deploymentModel.CreateParams) error {
	log.Println("setting up github for", name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup github for deployment %s. details: %s", name, err)
	}

	client, err := github.New(&github.ClientConf{
		Token: params.GitHub.Token,
	})

	if err != nil {
		return makeError(err)
	}

	deployment, err := deploymentModel.GetByName(name)
	if err != nil {
		return makeError(err)
	}

	if deployment == nil {
		return nil
	}

	var webhook *githubModels.WebhookPublic
	if deployment.Subsystems.GitHub.Webhook.ID == 0 {
		webhookTarget := fmt.Sprintf("%s/v1/hooks/deployments/github", conf.Env.ExternalUrl)
		id, err := client.CreateWebhook(&githubModels.WebhookPublic{
			RepositoryID: params.GitHub.RepositoryID,
			Events:       []string{"push"},
			Active:       true,
			ContentType:  "json",
			WebhookURL:   webhookTarget,
			Secret:       uuid.NewString(),
		})
		if err != nil {
			return makeError(err)
		}

		webhook, err = client.ReadWebhook(id, params.GitHub.RepositoryID)
		if err != nil {
			return makeError(err)
		}

		if webhook == nil {
			return makeError(fmt.Errorf("failed to read webhook after creation"))
		}

		err = deploymentModel.UpdateSubsystemByName(name, "github", "webhook", webhook)
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
		err := deploymentModel.UpdateSubsystemByName(name, "github", "webhook", githubModels.WebhookPublic{})
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

func CreateFakeGitHub(name string) error {
	log.Println("setting up placeholder github")

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup placeholder github. details: %s", err)
	}

	err := deploymentModel.UpdateSubsystemByName(name, "github", "webhook.id", 1)
	if err != nil {
		return makeError(err)
	}

	return nil
}