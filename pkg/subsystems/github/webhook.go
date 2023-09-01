package github

import (
	"context"
	"fmt"
	"github.com/google/go-github/github"
	"go-deploy/pkg/subsystems/github/models"
)

func (c *Client) ReadWebhook(id int64, repositoryID int64) (*models.WebhookPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read github webhook. details: %w", err)
	}

	repository, _, err := c.GitHubClient.Repositories.GetByID(context.TODO(), repositoryID)
	if err != nil {
		return nil, makeError(err)
	}

	if repository.Name == nil || repository.Owner == nil || repository.Owner.Login == nil {
		return nil, makeError(fmt.Errorf("failed to get repository name or owner"))
	}

	hook, _, err := c.GitHubClient.Repositories.GetHook(context.TODO(), *repository.Owner.Login, *repository.Name, id)
	if err != nil {
		return nil, makeError(err)
	}

	var public *models.WebhookPublic
	if hook != nil {
		public = models.CreateWebhookPublicFromGet(hook, repositoryID)
	}

	return public, nil
}

func (c *Client) CreateWebhook(public *models.WebhookPublic) (int64, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create github webhook. details: %w", err)
	}

	repository, _, err := c.GitHubClient.Repositories.GetByID(context.TODO(), public.RepositoryID)
	if err != nil {
		return 0, makeError(err)
	}

	if repository.Name == nil || repository.Owner == nil || repository.Owner.Login == nil {
		return 0, makeError(fmt.Errorf("failed to get repository name or owner"))
	}

	currentHooks, _, err := c.GitHubClient.Repositories.ListHooks(context.TODO(), *repository.Owner.Login, *repository.Name, nil)
	if err != nil {
		return 0, makeError(err)
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
		created, _, err := c.GitHubClient.Repositories.CreateHook(context.TODO(), *repository.Owner.Login, *repository.Name, &github.Hook{
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
			return 0, makeError(err)
		}

		if created.ID == nil {
			return 0, makeError(fmt.Errorf("failed to get webhook id after creation"))
		}

		id = created.ID
	}

	return *id, nil
}

func (c *Client) DeleteWebhook(id int64, repositoryId int64) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete github webhook. details: %w", err)
	}

	repository, _, err := c.GitHubClient.Repositories.GetByID(context.TODO(), repositoryId)
	if err != nil {
		return makeError(err)
	}

	if repository == nil {
		return nil
	}

	if repository.Name == nil || repository.Owner == nil || repository.Owner.Login == nil {
		return makeError(fmt.Errorf("failed to get repository name or owner"))
	}

	_, err = c.GitHubClient.Repositories.DeleteHook(context.TODO(), *repository.Owner.Login, *repository.Name, id)
	if err != nil {
		return makeError(err)
	}

	return nil
}
