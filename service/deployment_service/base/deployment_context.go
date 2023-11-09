package base

import (
	configModels "go-deploy/models/config"
	deploymentModels "go-deploy/models/sys/deployment"
	"go-deploy/pkg/config"
	"go-deploy/service/resources"
)

type DeploymentContext struct {
	Deployment *deploymentModels.Deployment
	MainApp    *deploymentModels.App
	Zone       *configModels.DeploymentZone
	Generator  *resources.PublicGeneratorType
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

func (dc *DeploymentContext) Refresh() error {
	deployment, err := deploymentModels.New().GetByID(dc.Deployment.ID)
	if err != nil {
		return err
	}

	if deployment == nil {
		return DeploymentDeletedErr
	}

	dc.Deployment = deployment
	dc.Generator.WithDeployment(deployment)
	return nil
}
