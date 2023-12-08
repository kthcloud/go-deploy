package github_service

import (
	"go-deploy/pkg/subsystems/github"
	"go-deploy/service/deployment_service/base"
	"go-deploy/service/resources"
)

type Context struct {
	base.DeploymentContext

	Client    *github.Client
	Generator *resources.GitHubGenerator
}

func NewContext(deploymentID, token string, repositoryID int64) (*Context, error) {
	baseContext, err := base.NewDeploymentBaseContext(deploymentID)
	if err != nil {
		return nil, err
	}

	client, err := withClient(token)
	if err != nil {
		return nil, err
	}

	return &Context{
		DeploymentContext: *baseContext,
		Client:            client,
		Generator:         baseContext.Generator.GitHub(token, repositoryID),
	}, nil
}

func withClient(token string) (*github.Client, error) {
	return github.New(&github.ClientConf{
		Token: token,
	})
}
