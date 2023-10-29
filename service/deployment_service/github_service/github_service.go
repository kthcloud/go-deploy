package github_service

import (
	"context"
	"errors"
	"fmt"
	githubThirdParty "github.com/google/go-github/github"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/pkg/conf"
	"go-deploy/service/resources"
	"go-deploy/service/vm_service/base"
	"log"
	"strings"
)

func Create(id string, params *deploymentModel.CreateParams) error {
	log.Println("setting up github for", params.Name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup github for deployment %s. details: %w", params.Name, err)
	}

	githubCtx, err := NewContext(id, params.GitHub.Token, params.GitHub.RepositoryID)
	if err != nil {
		if errors.Is(err, base.DeploymentDeletedErr) {
			return nil
		}

		return makeError(err)
	}

	githubCtx.WithCreateParams(params)

	// webhook
	err = resources.SsCreator(githubCtx.Client.CreateWebhook).
		WithDbFunc(dbFunc(id, "webhook")).
		WithPublic(githubCtx.Generator.Webhook()).
		Exec()

	return nil
}

func Delete(id string) error {
	log.Println("deleting github for", id)

	makeError := func(err error) error {
		return fmt.Errorf("failed to delete github for deployment %s. details: %w", id, err)
	}

	githubCtx, err := NewContext(id, "", 0)
	if err != nil {
		if errors.Is(err, base.DeploymentDeletedErr) {
			return nil
		}

		return makeError(err)
	}

	// webhook
	err = resources.SsDeleter(func(int64) error { return nil }).
		WithResourceID(githubCtx.Deployment.Subsystems.GitHub.Webhook.ID).
		WithDbFunc(dbFunc(id, "webhook")).
		Exec()

	if err != nil {
		return makeError(err)
	}

	return nil
}

func CreatePlaceholder(id string) error {
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

func ValidateToken(token string) (bool, string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to validate github token. details: %w", err)
	}

	client, err := withGitHubClient(token)
	if err != nil {
		return false, "", makeError(err)
	}

	limits, resp, err := client.GitHubClient.RateLimits(context.TODO())
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

func GetAccessTokenByCode(code string) (string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get github access token. details: %w", err)
	}

	token, prodErr := fetchAccessToken(code, conf.Env.GitHub.ProdClient.ID, conf.Env.GitHub.ProdClient.Secret)
	if prodErr == nil {
		return token, nil
	}

	token, devErr := fetchAccessToken(code, conf.Env.GitHub.DevClient.ID, conf.Env.GitHub.DevClient.Secret)
	if devErr == nil {
		return token, nil
	}

	return "", makeError(fmt.Errorf("failed to get github access token. prod err details: %w. dev err details: %w", prodErr, devErr))
}

func GetRepositories(token string) ([]deploymentModel.GitHubRepository, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get github repositories. details: %w", err)
	}

	client, err := withGitHubClient(token)
	if err != nil {
		return nil, makeError(err)
	}

	user, _, err := client.GitHubClient.Users.Get(context.Background(), "")
	if err != nil {
		return nil, makeError(err)
	}

	if user.Login == nil {
		return nil, makeError(fmt.Errorf("failed to get github repositories. user login is nil"))
	}

	repos, _, err := client.GitHubClient.Repositories.List(context.Background(), *user.Login, nil)
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

func GetRepository(token string, repositoryID int64) (*deploymentModel.GitHubRepository, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get repository. details: %w", err)
	}

	client, err := withGitHubClient(token)
	if err != nil {
		return nil, makeError(err)
	}

	repo, _, err := client.GitHubClient.Repositories.GetByID(context.Background(), repositoryID)
	if err != nil {
		return nil, makeError(err)
	}

	if repo.ID == nil || repo.Name == nil || repo.CloneURL == nil || repo.DefaultBranch == nil {
		log.Println("failed to get repository. one of the fields is nil. assuming it was deleted")
		return nil, nil
	}

	gitHubRepository := &deploymentModel.GitHubRepository{
		ID:            *repo.ID,
		Name:          *repo.Name,
		Owner:         *repo.Owner.Login,
		CloneURL:      *repo.CloneURL,
		DefaultBranch: *repo.DefaultBranch,
	}

	return gitHubRepository, nil
}

func GetWebhooks(token, owner, repository string) ([]deploymentModel.GitHubWebhook, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get repository webhooks. details: %w", err)
	}

	client, err := withGitHubClient(token)
	if err != nil {
		return nil, makeError(err)
	}

	webhooksRes, _, err := client.GitHubClient.Repositories.ListHooks(context.Background(), owner, repository, nil)
	if err != nil {
		return nil, makeError(err)
	}

	webhooks := make([]deploymentModel.GitHubWebhook, 0)
	for _, webhook := range webhooksRes {
		if webhook.ID != nil {
			webhooks = append(webhooks, deploymentModel.GitHubWebhook{
				ID:     *webhook.ID,
				Events: webhook.Events,
			})
		}
	}

	return webhooks, nil
}

func dbFunc(id, key string) func(interface{}) error {
	return func(data interface{}) error {
		if data == nil {
			return deploymentModel.New().DeleteSubsystemByID(id, "github."+key)
		}
		return deploymentModel.New().UpdateSubsystemByID(id, "github."+key, data)
	}
}
