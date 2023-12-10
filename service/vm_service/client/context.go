package client

import (
	"go-deploy/models/config"
	"go-deploy/models/sys/gpu"
	"go-deploy/models/sys/vm"
	"go-deploy/service"
	"go-deploy/service/resources"
)

type Context struct {
	vmStore  map[string]*vm.VM
	gpuStore map[string]*gpu.GPU

	UserID         string
	Zone           *config.VmZone
	DeploymentZone *config.DeploymentZone
	Generator      *resources.PublicGeneratorType
	Auth           *service.AuthInfo
}
