package harbor

import (
	"context"
	"fmt"
	modelv2 "github.com/mittwald/goharbor-client/v5/apiv2/model"
	"go-deploy/pkg/log"
	models "go-deploy/pkg/subsystems/harbor/models"
	"strings"
)

// ReadWebhook reads a webhook from Harbor.
func (client *Client) ReadWebhook(id int) (*models.WebhookPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read webhook for %d. details: %w", id, err)
	}

	project, err := client.HarborClient.GetProject(context.TODO(), client.Project)
	if err != nil {
		if strings.Contains(err.Error(), "project not found on server side") {
			log.Println("project", client.Project, "not found when deleting webhook. assuming it was deleted")
			return nil, nil
		}

		return nil, makeError(err)
	}

	webhookPolicies, err := client.HarborClient.ListProjectWebhookPolicies(context.TODO(), int(project.ProjectID))
	if err != nil {
		return nil, makeError(err)
	}

	for _, policy := range webhookPolicies {
		if int(policy.ID) == id {

			project, err := client.HarborClient.GetProject(context.TODO(), client.Project)
			if err != nil {
				return nil, makeError(err)
			}

			public := models.CreateWebhookPublicFromGet(policy, project)

			return public, nil
		}
	}
	return nil, nil
}

// CreateWebhook creates a webhook in Harbor.
func (client *Client) CreateWebhook(public *models.WebhookPublic) (*models.WebhookPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create webhook for %s. details: %w", public.Name, err)
	}

	project, err := client.HarborClient.GetProject(context.TODO(), client.Project)
	if err != nil {
		if strings.Contains(err.Error(), "project not found on server side") {
			log.Println("project", client.Project, "not found when deleting webhook. assuming it was deleted")
			return nil, nil
		}

		return nil, makeError(err)
	}

	webhookPolicies, err := client.HarborClient.ListProjectWebhookPolicies(context.TODO(), int(project.ProjectID))
	if err != nil {
		return nil, makeError(err)
	}

	var webhookPolicy *modelv2.WebhookPolicy
	for _, policy := range webhookPolicies {
		if len(policy.Targets) > 0 && policy.Targets[0].Address == public.Target {
			webhookPolicy = policy
		}
	}

	if webhookPolicy != nil {
		return models.CreateWebhookPublicFromGet(webhookPolicy, nil), nil
	}

	requestBody := models.CreateWebhookParamsFromPublic(public)
	err = client.HarborClient.AddProjectWebhookPolicy(context.TODO(), int(project.ProjectID), requestBody)
	if err != nil {
		return nil, makeError(err)
	}

	webhookPolicies, err = client.HarborClient.ListProjectWebhookPolicies(context.TODO(), int(project.ProjectID))
	if err != nil {
		return nil, makeError(err)
	}

	for _, policy := range webhookPolicies {
		if policy.Name == public.Name {
			return models.CreateWebhookPublicFromGet(policy, nil), nil
		}
	}

	return nil, makeError(fmt.Errorf("webhook not found after creation"))
}

// UpdateWebhook updates a webhook in Harbor.
func (client *Client) UpdateWebhook(public *models.WebhookPublic) (*models.WebhookPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update webhook for %d. details: %w", public.ID, err)
	}

	project, err := client.HarborClient.GetProject(context.TODO(), client.Project)
	if err != nil {
		if strings.Contains(err.Error(), "project not found on server side") {
			log.Println("project", client.Project, "not found when deleting webhook. assuming it was deleted")
			return nil, nil
		}

		return nil, makeError(err)
	}

	webhookPolicies, err := client.HarborClient.ListProjectWebhookPolicies(context.TODO(), int(project.ProjectID))
	if err != nil {
		return nil, makeError(err)
	}

	var webhookPolicy *modelv2.WebhookPolicy
	for _, policy := range webhookPolicies {
		if int(policy.ID) == public.ID {
			webhookPolicy = policy
		}
	}

	if webhookPolicy == nil {
		log.Println("webhook", public.Name, "not found when updating. assuming it was deleted")
		return nil, nil
	}

	params := models.CreateWebhookUpdateParamsFromPublic(public, webhookPolicy)
	err = client.HarborClient.UpdateProjectWebhookPolicy(context.TODO(), int(project.ProjectID), params)
	if err != nil {
		return nil, makeError(err)
	}

	webhookPolicies, err = client.HarborClient.ListProjectWebhookPolicies(context.TODO(), int(project.ProjectID))
	if err != nil {
		return nil, makeError(err)
	}

	for _, policy := range webhookPolicies {
		if policy.Name == public.Name {
			return models.CreateWebhookPublicFromGet(policy, nil), nil
		}
	}

	return nil, makeError(fmt.Errorf("webhook not found after update"))
}

// DeleteWebhook deletes a webhook from Harbor.
func (client *Client) DeleteWebhook(id int) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete webhook for %d. details: %w", id, err)
	}

	if id == 0 {
		log.Println("id not supplied when deleting webhook. assuming it was deleted")
		return nil
	}

	project, err := client.HarborClient.GetProject(context.TODO(), client.Project)
	if err != nil {
		if strings.Contains(err.Error(), "project not found on server side") {
			log.Println("project", client.Project, "not found when deleting webhook. assuming it was deleted")
			return nil
		}

		return makeError(err)
	}

	err = client.HarborClient.DeleteProjectWebhookPolicy(context.TODO(), int(project.ProjectID), int64(id))
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "[404] deleteWebhookPolicyOfProjectNotFound") {
			return nil
		}

		return makeError(err)
	}

	return nil
}
