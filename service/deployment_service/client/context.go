package client

import (
	deploymentModels "go-deploy/models/sys/deployment"
	"go-deploy/service"
)

type Context struct {
	deploymentStore map[string]*deploymentModels.Deployment

	Auth *service.AuthInfo
}
