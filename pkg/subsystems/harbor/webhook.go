package harbor

import (
	"context"
	"fmt"
	projectModels "go-deploy/pkg/imp/harbor/sdk/v2.0/client/project"
	webhookModels "go-deploy/pkg/imp/harbor/sdk/v2.0/client/webhook"
	harborModels "go-deploy/pkg/imp/harbor/sdk/v2.0/models"
	"go-deploy/pkg/log"
	models "go-deploy/pkg/subsystems/harbor/models"
	"strconv"
)

// ReadWebhook reads a webhook from Harbor.
func (client *Client) ReadWebhook(id int) (*models.WebhookPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read webhook for %d. details: %w", id, err)
	}

	project, err := client.HarborClient.V2().Project.GetProject(context.TODO(), &projectModels.GetProjectParams{ProjectNameOrID: client.Project})
	if err != nil {
		if IsNotFoundErr(err) {
			return nil, nil
		}

		return nil, makeError(err)
	}

	webhookPolicies, err := client.HarborClient.V2().Webhook.ListWebhookPoliciesOfProject(context.TODO(), &webhookModels.ListWebhookPoliciesOfProjectParams{
		ProjectNameOrID: strconv.Itoa(int(project.Payload.ProjectID)),
	})
	if err != nil {
		if IsNotFoundErr(err) {
			return nil, nil
		}

		return nil, makeError(err)
	}

	for _, policy := range webhookPolicies.Payload {
		if int(policy.ID) == id {
			project, err = client.HarborClient.V2().Project.GetProject(context.TODO(), &projectModels.GetProjectParams{
				ProjectNameOrID: strconv.Itoa(int(policy.ProjectID)),
			})
			if err != nil {
				if IsNotFoundErr(err) {
					return nil, nil
				}

				return nil, makeError(err)
			}

			public := models.CreateWebhookPublicFromGet(policy, project.Payload)

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

	project, err := client.HarborClient.V2().Project.GetProject(context.TODO(), &projectModels.GetProjectParams{ProjectNameOrID: client.Project})
	if err != nil {
		if IsNotFoundErr(err) {
			return nil, nil
		}

		return nil, makeError(err)
	}

	webhookPolicies, err := client.HarborClient.V2().Webhook.ListWebhookPoliciesOfProject(context.TODO(), &webhookModels.ListWebhookPoliciesOfProjectParams{
		ProjectNameOrID: strconv.Itoa(int(project.Payload.ProjectID)),
	})
	if err != nil {
		if !IsNotFoundErr(err) {
			return nil, makeError(err)
		}
	}

	var webhookPolicy *harborModels.WebhookPolicy
	if webhookPolicies != nil {
		for _, policy := range webhookPolicies.Payload {
			if len(policy.Targets) > 0 && policy.Targets[0].Address == public.Target {
				webhookPolicy = policy
			}
		}
	}

	if webhookPolicy != nil {
		return models.CreateWebhookPublicFromGet(webhookPolicy, nil), nil
	}

	requestBody := models.CreateWebhookParamsFromPublic(public)
	_, err = client.HarborClient.V2().Webhook.CreateWebhookPolicyOfProject(context.TODO(), &webhookModels.CreateWebhookPolicyOfProjectParams{
		Policy:          requestBody,
		ProjectNameOrID: strconv.Itoa(int(project.Payload.ProjectID)),
	})
	if err != nil {
		if IsNotFoundErr(err) {
			return nil, nil
		}

		return nil, makeError(err)
	}

	webhookPolicies, err = client.HarborClient.V2().Webhook.ListWebhookPoliciesOfProject(context.TODO(), &webhookModels.ListWebhookPoliciesOfProjectParams{
		ProjectNameOrID: strconv.Itoa(int(project.Payload.ProjectID)),
	})
	if err != nil {
		if IsNotFoundErr(err) {
			return nil, nil
		}

		return nil, makeError(err)
	}

	if webhookPolicies != nil {
		for _, policy := range webhookPolicies.Payload {
			if policy.Name == public.Name {
				return models.CreateWebhookPublicFromGet(policy, nil), nil
			}
		}
	}

	return nil, makeError(fmt.Errorf("webhook not found after creation"))
}

// UpdateWebhook updates a webhook in Harbor.
func (client *Client) UpdateWebhook(public *models.WebhookPublic) (*models.WebhookPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update webhook for %d. details: %w", public.ID, err)
	}

	project, err := client.HarborClient.V2().Project.GetProject(context.TODO(), &projectModels.GetProjectParams{ProjectNameOrID: client.Project})
	if err != nil {
		if IsNotFoundErr(err) {
			return nil, nil
		}

		return nil, makeError(err)
	}

	webhookPolicies, err := client.HarborClient.V2().Webhook.ListWebhookPoliciesOfProject(context.TODO(), &webhookModels.ListWebhookPoliciesOfProjectParams{
		ProjectNameOrID: strconv.Itoa(int(project.Payload.ProjectID)),
	})
	if err != nil {
		if IsNotFoundErr(err) {
			return nil, nil
		}

		return nil, makeError(err)
	}

	var webhookPolicy *harborModels.WebhookPolicy
	if webhookPolicies != nil {
		for _, policy := range webhookPolicies.Payload {
			if int(policy.ID) == public.ID {
				webhookPolicy = policy
			}
		}
	}

	if webhookPolicy == nil {
		log.Println("webhook", public.Name, "not found when updating. Assuming it was deleted")
		return nil, nil
	}

	params := models.CreateWebhookUpdateParamsFromPublic(public, webhookPolicy)
	_, err = client.HarborClient.V2().Webhook.UpdateWebhookPolicyOfProject(context.TODO(), &webhookModels.UpdateWebhookPolicyOfProjectParams{
		Policy:          params,
		ProjectNameOrID: strconv.Itoa(int(project.Payload.ProjectID)),
		WebhookPolicyID: int64(public.ID),
	})
	if err != nil {
		if IsNotFoundErr(err) {
			return nil, nil
		}

		return nil, makeError(err)
	}

	webhookPolicies, err = client.HarborClient.V2().Webhook.ListWebhookPoliciesOfProject(context.TODO(), &webhookModels.ListWebhookPoliciesOfProjectParams{
		ProjectNameOrID: strconv.Itoa(int(project.Payload.ProjectID)),
	})
	if err != nil {
		if IsNotFoundErr(err) {
			return nil, nil
		}

		return nil, makeError(err)
	}

	if webhookPolicies != nil {
		for _, policy := range webhookPolicies.Payload {
			if policy.Name == public.Name {
				return models.CreateWebhookPublicFromGet(policy, nil), nil
			}
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
		log.Println("ID not supplied when deleting webhook. Assuming it was deleted")
		return nil
	}

	project, err := client.HarborClient.V2().Project.GetProject(context.TODO(), &projectModels.GetProjectParams{ProjectNameOrID: client.Project})
	if err != nil {
		if IsNotFoundErr(err) {
			return nil
		}

		return makeError(err)
	}

	_, err = client.HarborClient.V2().Webhook.DeleteWebhookPolicyOfProject(context.TODO(), &webhookModels.DeleteWebhookPolicyOfProjectParams{
		ProjectNameOrID: strconv.Itoa(int(project.Payload.ProjectID)),
		WebhookPolicyID: int64(id),
	})
	if err != nil {
		if IsNotFoundErr(err) {
			return nil
		}

		return makeError(err)
	}

	return nil
}
