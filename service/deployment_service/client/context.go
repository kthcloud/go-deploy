package client

import (
	configModels "go-deploy/models/config"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/service"
	"go-deploy/service/resources"
)

type Context struct {
	id     string
	IDs    []string
	name   string
	UserID string

	deployment *deploymentModel.Deployment
	MainApp    *deploymentModel.App
	zone       *configModels.DeploymentZone
	Generator  *resources.PublicGeneratorType

	Auth *service.AuthInfo
}
