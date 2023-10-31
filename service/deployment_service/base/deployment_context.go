package base

import (
	configModels "go-deploy/models/config"
	deploymentModels "go-deploy/models/sys/deployment"
	"go-deploy/pkg/config"
	"go-deploy/service/resources"
)

type DeploymentContext struct {
	Deployment   *deploymentModels.Deployment
	MainApp      *deploymentModels.App
	Zone         *configModels.DeploymentZone
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

	zone := config.Config.Deployment.GetZone(deployment.Zone)
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
