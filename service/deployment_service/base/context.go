package base

import (
	deploymentModels "go-deploy/models/sys/deployment"
	"go-deploy/models/sys/enviroment"
	"go-deploy/pkg/conf"
)

type Context struct {
	Deployment *deploymentModels.Deployment
	MainApp    *deploymentModels.App
	Zone       *enviroment.DeploymentZone

	CreateParams *deploymentModels.CreateParams
	UpdateParams *deploymentModels.UpdateParams
}

func NewBaseContext(deploymentID string) (*Context, error) {
	deployment, err := deploymentModels.New().GetByID(deploymentID)
	if err != nil {
		return nil, err
	}

	if deployment == nil {
		return nil, DeploymentDeletedErr
	}

	mainApp := deployment.GetMainApp()
	if mainApp == nil {
		return nil, MainAppNotFoundErr
	}

	zone := conf.Env.Deployment.GetZone(deployment.Zone)
	if zone == nil {
		return nil, ZoneNotFoundErr
	}

	return &Context{
		Deployment: deployment,
		MainApp:    mainApp,
		Zone:       zone,
	}, nil
}

func (c *Context) WithCreateParams(params *deploymentModels.CreateParams) *Context {
	c.CreateParams = params
	return c
}

func (c *Context) WithUpdateParams(params *deploymentModels.UpdateParams) *Context {
	c.UpdateParams = params
	return c
}
