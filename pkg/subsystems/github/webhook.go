package github

import (
	"context"
	"fmt"
	"github.com/google/go-github/github"
	"go-deploy/pkg/subsystems/github/models"
)

// ListWebhooks lists all webhooks for a repository.
func (client *Client) ListWebhooks(owner string, repositoryName string) ([]*models.WebhookPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to list github webhooks. details: %w", err)
	}

	hooks, _, err := client.GitHubClient.Repositories.ListHooks(context.TODO(), owner, repositoryName, nil)
	if err != nil {
		return nil, makeError(err)
	}

	var public []*models.WebhookPublic
	for _, hook := range hooks {
		if hook.Config == nil {
			continue
		}

		public = append(public, models.CreateWebhookPublicFromGet(hook, 0))
	}

	return public, nil
}

// ReadWebhook reads a webhook from GitHub.
func (client *Client) ReadWebhook(id int64, repositoryID int64) (*models.WebhookPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read github webhook. details: %w", err)
	}

	repository, _, err := client.GitHubClient.Repositories.GetByID(context.TODO(), repositoryID)
	if err != nil {
		return nil, makeError(err)
	}

	if repository.Name == nil || repository.Owner == nil || repository.Owner.Login == nil {
		return nil, makeError(fmt.Errorf("failed to get repository name or owner"))
	}

	hook, _, err := client.GitHubClient.Repositories.GetHook(context.TODO(), *repository.Owner.Login, *repository.Name, id)
	if err != nil {
		return nil, makeError(err)
	}

	var public *models.WebhookPublic
	if hook != nil {
		public = models.CreateWebhookPublicFromGet(hook, repositoryID)
	}

	return public, nil
}

// CreateWebhook creates a webhook on GitHub.
func (client *Client) CreateWebhook(public *models.WebhookPublic) (*models.WebhookPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create github webhook. details: %w", err)
	}

	repository, _, err := client.GitHubClient.Repositories.GetByID(context.TODO(), public.RepositoryID)
	if err != nil {
		return nil, makeError(err)
	}

	if repository.Name == nil || repository.Owner == nil || repository.Owner.Login == nil {
		return nil, makeError(fmt.Errorf("failed to get repository name or owner"))
	}

	currentHooks, _, err := client.GitHubClient.Repositories.ListHooks(context.TODO(), *repository.Owner.Login, *repository.Name, nil)
	if err != nil {
		return nil, makeError(err)
	}

	var id *int64
	for _, hook := range currentHooks {
		if hook.Config == nil {
			continue
		}

		if hook.Config["url"] == public.WebhookURL {
			id = hook.ID
			break
		}
	}

	if id == nil {
		created, _, err := client.GitHubClient.Repositories.CreateHook(context.TODO(), *repository.Owner.Login, *repository.Name, &github.Hook{
			Name:   github.String("web"),
			Active: github.Bool(true),
			Events: []string{"push"},
			Config: map[string]interface{}{
				"url":          public.WebhookURL,
				"content_type": "json",
				"insecure_ssl": false,
				"secret":       public.Secret,
				"token":        public.Secret,
			},
		})

		if err != nil {
			return nil, makeError(err)
		}

		if created.ID == nil {
			return nil, makeError(fmt.Errorf("failed to get webhook id after creation"))
		}

		id = created.ID
	}

	return client.ReadWebhook(*id, public.RepositoryID)
}

// DeleteWebhook deletes a webhook from GitHub.
func (client *Client) DeleteWebhook(id int64, repositoryId int64) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete github webhook. details: %w", err)
	}

	repository, _, err := client.GitHubClient.Repositories.GetByID(context.TODO(), repositoryId)
	if err != nil {
		return makeError(err)
	}

	if repository == nil {
		return nil
	}

	if repository.Name == nil || repository.Owner == nil || repository.Owner.Login == nil {
		return makeError(fmt.Errorf("failed to get repository name or owner"))
	}

	_, err = client.GitHubClient.Repositories.DeleteHook(context.TODO(), *repository.Owner.Login, *repository.Name, id)
	if err != nil {
		return makeError(err)
	}

	return nil
}
