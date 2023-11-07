package harbor_service

import (
	"go-deploy/pkg/config"
	"go-deploy/pkg/subsystems/harbor"
	"go-deploy/service/deployment_service/base"
	"go-deploy/service/resources"
	"go-deploy/utils/subsystemutils"
)

type Context struct {
	base.DeploymentContext

	Client    *harbor.Client
	Generator *resources.HarborGenerator
}

func NewContext(deploymentID string, ownerID ...string) (*Context, error) {
	baseContext, err := base.NewDeploymentBaseContext(deploymentID)
	if err != nil {
		return nil, err
	}

	var projectName string
	if len(ownerID) > 0 {
		projectName = getProjectName(ownerID[0])
	} else {
		projectName = getProjectName(baseContext.Deployment.OwnerID)
	}

	harborClient, err := withClient(projectName)
	if err != nil {
		return nil, err
	}

	return &Context{
		DeploymentContext: *baseContext,
		Client:            harborClient,
		Generator:         baseContext.Generator.Harbor(harborClient.Project),
	}, nil
}

func getProjectName(userID string) string {
	return subsystemutils.GetPrefixedName(userID)
}

func withClient(project string) (*harbor.Client, error) {
	return harbor.New(&harbor.ClientConf{
		URL:      config.Config.Harbor.URL,
		Username: config.Config.Harbor.User,
		Password: config.Config.Harbor.Password,
		Project:  project,
	})
}
