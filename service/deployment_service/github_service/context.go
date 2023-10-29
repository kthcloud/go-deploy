package github_service

import (
	"go-deploy/pkg/subsystems/github"
	"go-deploy/service/deployment_service/base"
	"go-deploy/service/resources"
	"go-deploy/utils/subsystemutils"
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

func getProjectName(userID string) string {
	return subsystemutils.GetPrefixedName(userID)
}

func withClient(token string) (*github.Client, error) {
	return github.New(&github.ClientConf{
		Token: token,
	})
}
