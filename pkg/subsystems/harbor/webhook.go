package harbor

import (
	"context"
	"fmt"
	modelv2 "github.com/mittwald/goharbor-client/v5/apiv2/model"
	"go-deploy/utils"
)

func (client *Client) WebhookCreated(projectName, name string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if harbor repository %s is created. details: %s", name, err)
	}

	project, err := client.HarborClient.GetProject(context.TODO(), projectName)
	if err != nil {
		return false, makeError(err)
	}

	if project == nil {
		return false, nil
	}

	webhookPolicies, err := client.HarborClient.ListProjectWebhookPolicies(context.TODO(), int(project.ProjectID))
	if err != nil {
		return false, makeError(err)
	}

	for _, policy := range webhookPolicies {
		if policy.Name == name {
			return true, nil
		}
	}
	return false, nil
}

func (client *Client) CreateWebhook(projectName, name, webhookTarget string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create harbor webhook for %s. details: %s", name, err)
	}

	projectExists, project, err := client.assertProjectExists(projectName)
	if err != nil {
		return makeError(err)
	}

	if !projectExists {
		err = fmt.Errorf("no project exists")
		return makeError(err)
	}

	webhooks, err := client.HarborClient.ListProjectWebhookPolicies(context.TODO(), int(project.ProjectID))
	if err != nil {
		return makeError(err)
	}

	for _, hook := range webhooks {
		if hook.Name == name {
			return nil
		}
	}

	webhookToken, err := generateToken(client.webhookSecret)
	if err != nil {
		return makeError(err)
	}

	err = updateDatabaseWebhook(name, utils.HashString(webhookToken))
	if err != nil {
		return makeError(err)
	}

	err = client.HarborClient.AddProjectWebhookPolicy(context.TODO(), int(project.ProjectID), &modelv2.WebhookPolicy{
		Enabled:    true,
		EventTypes: getWebhookEventTypes(),
		Name:       name,
		Targets: []*modelv2.WebhookTargetObject{
			{
				Address:        webhookTarget,
				AuthHeader:     createAuthHeader(webhookToken),
				SkipCertVerify: false,
				Type:           "http",
			},
		},
	})

	if err != nil {
		return makeError(err)
	}

	return nil
}
