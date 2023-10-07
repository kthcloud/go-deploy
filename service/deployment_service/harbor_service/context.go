package harbor_service

import (
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/harbor"
	"go-deploy/service/deployment_service/base"
	"go-deploy/service/deployment_service/resources"
	"go-deploy/utils/subsystemutils"
)

type Context struct {
	base.Context

	Client    *harbor.Client
	Generator *resources.HarborGenerator
}

func NewContext(deploymentID string) (*Context, error) {
	baseContext, err := base.NewBaseContext(deploymentID)
	if err != nil {
		return nil, err
	}

	harborClient, err := withClient(getProjectName(baseContext.Deployment.OwnerID))
	if err != nil {
		return nil, err
	}

	return &Context{
		BaseContext: *baseContext,
		Client:      harborClient,
		Generator:   resources.PublicGenerator().WithDeployment(baseContext.Deployment).Harbor(harborClient.Project),
	}, nil
}

func getProjectName(userID string) string {
	return subsystemutils.GetPrefixedName(userID)
}

func withClient(project string) (*harbor.Client, error) {
	return harbor.New(&harbor.ClientConf{
		URL:      conf.Env.Harbor.URL,
		Username: conf.Env.Harbor.User,
		Password: conf.Env.Harbor.Password,
		Project:  project,
	})
}
