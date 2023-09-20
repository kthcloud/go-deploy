package harbor

import (
	"context"
	"fmt"
	modelv2 "github.com/mittwald/goharbor-client/v5/apiv2/model"
	models "go-deploy/pkg/subsystems/harbor/models"
	"log"
	"strconv"
	"strings"
)

func (client *Client) WebhookCreated(public *models.WebhookPublic) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if webhook %s is created. details: %w", public.Name, err)
	}

	webhookPolicies, err := client.HarborClient.ListProjectWebhookPolicies(context.TODO(), public.ProjectID)
	if err != nil {
		return false, makeError(err)
	}

	for _, policy := range webhookPolicies {
		if int(policy.ID) == public.ID {
			return true, nil
		}
	}
	return false, nil
}

func (client *Client) ReadWebhook(projectID, id int) (*models.WebhookPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read webhook for %d. details: %w", id, err)
	}

	webhookPolicies, err := client.HarborClient.ListProjectWebhookPolicies(context.TODO(), projectID)
	if err != nil {
		return nil, makeError(err)
	}

	for _, policy := range webhookPolicies {
		if int(policy.ID) == id {

			project, err := client.HarborClient.GetProject(context.TODO(), strconv.Itoa(projectID))
			if err != nil {
				return nil, makeError(err)
			}

			public := models.CreateWebhookPublicFromGet(policy, project)

			return public, nil
		}
	}
	return nil, nil
}

func (client *Client) CreateWebhook(public *models.WebhookPublic) (int, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create webhook for %s. details: %w", public.Name, err)
	}

	if public.ProjectID == 0 {
		return 0, makeError(fmt.Errorf("project id required"))
	}

	webhookPolicies, err := client.HarborClient.ListProjectWebhookPolicies(context.TODO(), public.ProjectID)
	if err != nil {
		return 0, makeError(err)
	}

	var webhookPolicy *modelv2.WebhookPolicy
	for _, policy := range webhookPolicies {
		if len(policy.Targets) > 0 && policy.Targets[0].Address == public.Target {
			webhookPolicy = policy
		}
	}

	if webhookPolicy != nil {
		return int(webhookPolicy.ID), nil
	}

	requestBody := models.CreateWebhookParamsFromPublic(public)
	err = client.HarborClient.AddProjectWebhookPolicy(context.TODO(), public.ProjectID, requestBody)
	if err != nil {
		return 0, makeError(err)
	}

	webhookPolicies, err = client.HarborClient.ListProjectWebhookPolicies(context.TODO(), public.ProjectID)
	if err != nil {
		return 0, makeError(err)
	}

	for _, policy := range webhookPolicies {
		if policy.Name == public.Name {
			return int(policy.ID), nil
		}
	}
	return 0, makeError(fmt.Errorf("webhook not found after creation"))
}

func (client *Client) UpdateWebhook(public *models.WebhookPublic) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update webhook for %s. details: %w", public.Name, err)
	}

	if public.ProjectID == 0 {
		return makeError(fmt.Errorf("project id required"))
	}

	webhookPolicies, err := client.HarborClient.ListProjectWebhookPolicies(context.TODO(), public.ProjectID)
	if err != nil {
		return makeError(err)
	}

	var webhookPolicy *modelv2.WebhookPolicy
	for _, policy := range webhookPolicies {
		if int(policy.ID) == public.ID {
			webhookPolicy = policy
		}
	}

	if webhookPolicy == nil {
		log.Println("webhook", public.Name, "not found when updating. assuming it was deleted")
		return nil
	}

	params := models.CreateWebhookUpdateParamsFromPublic(public, webhookPolicy)
	err = client.HarborClient.UpdateProjectWebhookPolicy(context.TODO(), public.ProjectID, params)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func (client *Client) DeleteWebhook(projectID, id int) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete webhook for %d. details: %w", id, err)
	}

	err := client.HarborClient.DeleteProjectWebhookPolicy(context.TODO(), projectID, int64(id))
	if err != nil {
		errMsg := err.Error()
		if strings.Contains(errMsg, "[404] deleteWebhookPolicyOfProjectNotFound") {
			return nil
		}

		return makeError(err)
	}

	return nil
}

func (client *Client) DeleteAllWebhooks(projectID int) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete all webhook for %d. details: %w", projectID, err)
	}

	webhookPolicies, err := client.HarborClient.ListProjectWebhookPolicies(context.TODO(), projectID)
	if err != nil {
		return makeError(err)
	}

	for _, policy := range webhookPolicies {
		err = client.HarborClient.DeleteProjectWebhookPolicy(context.TODO(), projectID, int64(policy.ID))
		if err != nil {
			errMsg := err.Error()
			if strings.Contains(errMsg, "[404] deleteWebhookPolicyOfProjectNotFound") {
				continue
			}

			return makeError(err)
		}
	}

	return nil
}
