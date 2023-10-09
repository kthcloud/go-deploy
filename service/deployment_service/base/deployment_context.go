package base

import (
	deploymentModels "go-deploy/models/sys/deployment"
	"go-deploy/models/sys/enviroment"
	"go-deploy/pkg/conf"
	"go-deploy/service/resources"
)

type DeploymentContext struct {
	Deployment   *deploymentModels.Deployment
	MainApp      *deploymentModels.App
	Zone         *enviroment.DeploymentZone
	CreateParams *deploymentModels.CreateParams
	UpdateParams *deploymentModels.UpdateParams
	Generator    *resources.PublicGeneratorType
}

func NewDeploymentBaseContext(deploymentID string) (*DeploymentContext, error) {
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

	return &DeploymentContext{
		Deployment: deployment,
		MainApp:    mainApp,
		Zone:       zone,
		Generator:  resources.PublicGenerator().WithDeploymentZone(zone).WithDeployment(deployment),
	}, nil
}

func (c *DeploymentContext) WithCreateParams(params *deploymentModels.CreateParams) *DeploymentContext {
	c.CreateParams = params
	return c
}

func (c *DeploymentContext) WithUpdateParams(params *deploymentModels.UpdateParams) *DeploymentContext {
	c.UpdateParams = params
	return c
}
