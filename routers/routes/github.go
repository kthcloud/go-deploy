package routes

import "go-deploy/routers/api/v1/v1_github"

const (
	GitHubRepositoriesPath = "/v1/github/repositories"
)

type GitHubRoutingGroup struct{ RoutingGroupBase }

func GitHubRoutes() *GitHubRoutingGroup {
	return &GitHubRoutingGroup{}
}

func (group *GitHubRoutingGroup) PrivateRoutes() []Route {
	return []Route{
		{Method: "GET", Pattern: GitHubRepositoriesPath, HandlerFunc: v1_github.ListGitHubRepositories},
	}
}
