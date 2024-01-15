package github_service

import (
	"context"
	"errors"
	"fmt"
	githubThirdParty "github.com/google/go-github/github"
	deploymentModels "go-deploy/models/sys/deployment"
	"go-deploy/pkg/config"
	"go-deploy/service/resources"
	"log"
	"strings"
)

// Create sets up the GitHub setup for the deployment.
// It creates a webhook associated with the deployment and returns an error if any.
func (c *Client) Create(id string, params *deploymentModels.CreateParams) error {
	log.Println("setting up github for", params.Name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup github for deployment %s. details: %w", params.Name, err)
	}

	_, gc, g, err := c.Get(OptsAll(id))
	if err != nil {
		return makeError(err)
	}

	// webhook
	err = resources.SsCreator(gc.CreateWebhook).
		WithDbFunc(dbFunc(id, "webhook")).
		WithPublic(g.Webhook()).
		Exec()

	if err != nil {
		return makeError(err)
	}

	return nil
}

// Delete deletes the GitHub setup for the deployment.
// It deletes the webhook associated with the deployment and returns an error if any.
func (c *Client) Delete(id string) error {
	log.Println("deleting github for", id)

	makeError := func(err error) error {
		return fmt.Errorf("failed to delete github for deployment %s. details: %w", id, err)
	}

	d, _, _, err := c.Get(OptsOnlyDeployment(id))
	if err != nil {
		return makeError(err)
	}

	// webhook
	err = resources.SsDeleter(func(int64) error { return nil }).
		WithResourceID(d.Subsystems.GitHub.Webhook.ID).
		WithDbFunc(dbFunc(id, "webhook")).
		Exec()

	if err != nil {
		return makeError(err)
	}

	return nil
}

// CreatePlaceholder sets up the placeholder GitHub for the deployment.
// This is intended to make the deployment aware GitHub has been set up for it, without actually setting it up.
func (c *Client) CreatePlaceholder(id string) error {
	log.Println("setting up placeholder github")

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup placeholder harbor. details: %w", err)
	}

	err := resources.SsPlaceholderCreator().WithDbFunc(dbFunc(id, "placeholder")).Exec()
	if err != nil {
		return makeError(err)
	}

	return nil
}

// Validate checks the validity of the GitHub token.
//
// It validates the token by checking the rate limits, response status code,
// core request limits, and required scopes.
//
// It returns true if the token is valid, along with an empty string and nil error.
// If the token is invalid, it returns false, an error message, and nil error.
// If any error occurs during validation, it returns false, an empty string, and the error.
func (c *Client) Validate() (bool, string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to validate github token. details: %w", err)
	}

	_, gc, _, err := c.Get(OptsOnlyClient())
	if err != nil {
		return false, "", makeError(err)
	}

	limits, resp, err := gc.GitHubClient.RateLimits(context.TODO())
	if err != nil {
		var githubError *githubThirdParty.ErrorResponse
		if errors.As(err, &githubError) && githubError.Message == "Bad credentials" {
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
		return false, "Requires scope to be one of: admin:repo_hook", nil
	}

	return true, "", nil
}

// GetRepositories gets the GitHub repositories associated with the token.
//
// It returns a list of GitHub repositories and nil error if any.
// If any error occurs, it returns nil and the error.
func (c *Client) GetRepositories() ([]deploymentModels.GitHubRepository, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get github repositories. details: %w", err)
	}

	_, gc, _, err := c.Get(OptsOnlyClient())
	if err != nil {
		return nil, makeError(err)
	}

	user, _, err := gc.GitHubClient.Users.Get(context.Background(), "")
	if err != nil {
		return nil, makeError(err)
	}

	if user.Login == nil {
		return nil, makeError(fmt.Errorf("failed to get github repositories. user login is nil"))
	}

	repos, _, err := gc.GitHubClient.Repositories.List(context.Background(), *user.Login, nil)
	if err != nil {
		return nil, makeError(err)
	}

	gitHubRepos := make([]deploymentModels.GitHubRepository, 0)
	for _, repo := range repos {
		if repo.ID == nil || repo.Name == nil || repo.CloneURL == nil || repo.DefaultBranch == nil {
			continue
		}

		gitHubRepos = append(gitHubRepos, deploymentModels.GitHubRepository{
			ID:            *repo.ID,
			Name:          *repo.Name,
			Owner:         *user.Login,
			CloneURL:      *repo.CloneURL,
			DefaultBranch: *repo.DefaultBranch,
		})
	}

	return gitHubRepos, nil
}

// GetRepository retrieves a GitHub repository
//
// Uses the ID from WithRepositoryID to retrieve the repository.
func (c *Client) GetRepository() (*deploymentModels.GitHubRepository, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get repository. details: %w", err)
	}

	_, gc, _, err := c.Get(OptsOnlyClient())
	if err != nil {
		return nil, makeError(err)
	}

	repo, err := gc.ReadRepository(c.repositoryID)
	if err != nil {
		return nil, makeError(err)
	}

	if repo == nil {
		return nil, nil
	}

	return &deploymentModels.GitHubRepository{
		ID:            repo.ID,
		Name:          repo.Name,
		Owner:         repo.Owner,
		CloneURL:      repo.CloneURL,
		DefaultBranch: repo.DefaultBranch,
	}, nil
}

// GetWebhooks retrieves a list of GitHub webhooks
//
// It uses the repository from WithRepository to retrieve the webhooks.
func (c *Client) GetWebhooks(repository *deploymentModels.GitHubRepository) ([]deploymentModels.GitHubWebhook, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get repository webhooks. details: %w", err)
	}

	_, gc, _, err := c.Get(OptsOnlyClient())
	if err != nil {
		return nil, makeError(err)
	}

	webhooks, err := gc.ListWebhooks(repository.Owner, repository.Name)
	if err != nil {
		return nil, makeError(err)
	}

	res := make([]deploymentModels.GitHubWebhook, len(webhooks))
	for idx, webhook := range webhooks {
		res[idx] = deploymentModels.GitHubWebhook{
			ID:     webhook.ID,
			Events: webhook.Events,
		}
	}

	return res, nil
}

// GetAccessTokenByCode retrieves the GitHub access token by the code.
//
// It returns the access token and nil error if any.
// If any error occurs, it returns an empty string and the error.
func GetAccessTokenByCode(code string) (string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get github access token. details: %w", err)
	}

	token, prodErr := fetchAccessToken(code, config.Config.GitHub.ProdClient.ID, config.Config.GitHub.ProdClient.Secret)
	if prodErr == nil {
		return token, nil
	}

	token, devErr := fetchAccessToken(code, config.Config.GitHub.DevClient.ID, config.Config.GitHub.DevClient.Secret)
	if devErr == nil {
		return token, nil
	}

	return "", makeError(fmt.Errorf("failed to get github access token. prod err details: %w. dev err details: %w", prodErr, devErr))
}

// dbFunc returns a function that updates the GitHub subsystem.
func dbFunc(id, key string) func(interface{}) error {
	return func(data interface{}) error {
		if data == nil {
			return deploymentModels.New().DeleteSubsystem(id, "github."+key)
		}
		return deploymentModels.New().SetSubsystem(id, "github."+key, data)
	}
}
