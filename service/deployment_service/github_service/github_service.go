package github_service

import (
	"context"
	"errors"
	"fmt"
	githubThirdParty "github.com/google/go-github/github"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/pkg/config"
	"go-deploy/service/resources"
	"log"
	"strings"
)

func (c *Client) Create(params *deploymentModel.CreateParams) error {
	log.Println("setting up github for", params.Name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup github for deployment %s. details: %w", params.Name, err)
	}

	_, gc, g, err := c.Get(OptsNoDeployment)
	if err != nil {
		return makeError(err)
	}

	// webhook
	err = resources.SsCreator(gc.CreateWebhook).
		WithDbFunc(dbFunc(c.ID(), "webhook")).
		WithPublic(g.Webhook()).
		Exec()

	return nil
}

func (c *Client) Delete() error {
	log.Println("deleting github for", c.ID())

	makeError := func(err error) error {
		return fmt.Errorf("failed to delete github for deployment %s. details: %w", c.ID(), err)
	}

	d, _, _, err := c.Get(OptsOnlyDeployment)
	if err != nil {
		return makeError(err)
	}

	// webhook
	err = resources.SsDeleter(func(int64) error { return nil }).
		WithResourceID(d.Subsystems.GitHub.Webhook.ID).
		WithDbFunc(dbFunc(c.ID(), "webhook")).
		Exec()

	if err != nil {
		return makeError(err)
	}

	return nil
}

func (c *Client) CreatePlaceholder() error {
	log.Println("setting up placeholder github")

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup placeholder harbor. details: %w", err)
	}

	err := resources.SsPlaceholderCreator().WithDbFunc(dbFunc(c.ID(), "placeholder")).Exec()
	if err != nil {
		return makeError(err)
	}

	return nil
}

func (c *Client) Validate() (bool, string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to validate github token. details: %w", err)
	}

	_, gc, _, err := c.Get(OptsOnlyClient)
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

func (c *Client) GetRepositories() ([]deploymentModel.GitHubRepository, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get github repositories. details: %w", err)
	}

	_, gc, _, err := c.Get(OptsOnlyClient)
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

	gitHubRepos := make([]deploymentModel.GitHubRepository, 0)
	for _, repo := range repos {
		if repo.ID == nil || repo.Name == nil || repo.CloneURL == nil || repo.DefaultBranch == nil {
			continue
		}

		gitHubRepos = append(gitHubRepos, deploymentModel.GitHubRepository{
			ID:            *repo.ID,
			Name:          *repo.Name,
			Owner:         *user.Login,
			CloneURL:      *repo.CloneURL,
			DefaultBranch: *repo.DefaultBranch,
		})
	}

	return gitHubRepos, nil
}

func (c *Client) GetRepository() (*deploymentModel.GitHubRepository, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get repository. details: %w", err)
	}

	_, gc, _, err := c.Get(OptsOnlyClient)
	if err != nil {
		return nil, makeError(err)
	}

	repo, err := gc.ReadRepository(c.RepositoryID())
	if err != nil {
		return nil, makeError(err)
	}

	if repo == nil {
		return nil, nil
	}

	return &deploymentModel.GitHubRepository{
		ID:            repo.ID,
		Name:          repo.Name,
		Owner:         repo.Owner,
		CloneURL:      repo.CloneURL,
		DefaultBranch: repo.DefaultBranch,
	}, nil
}

func (c *Client) GetWebhooks() ([]deploymentModel.GitHubWebhook, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get repository webhooks. details: %w", err)
	}

	_, gc, _, err := c.Get(OptsOnlyClient)
	if err != nil {
		return nil, makeError(err)
	}

	r := c.Repository()
	if r == nil {
		return nil, nil
	}

	webhooks, err := gc.ListWebhooks(r.Owner, r.Name)
	if err != nil {
		return nil, makeError(err)
	}

	res := make([]deploymentModel.GitHubWebhook, len(webhooks))
	for idx, webhook := range webhooks {
		res[idx] = deploymentModel.GitHubWebhook{
			ID:     webhook.ID,
			Events: webhook.Events,
		}
	}

	return res, nil
}

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

func dbFunc(id, key string) func(interface{}) error {
	return func(data interface{}) error {
		if data == nil {
			return deploymentModel.New().DeleteSubsystemByID(id, "github."+key)
		}
		return deploymentModel.New().UpdateSubsystemByID(id, "github."+key, data)
	}
}
