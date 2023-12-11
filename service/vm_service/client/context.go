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

	Generator *resources.PublicGeneratorType
	Auth      *service.AuthInfo

	// The following fields are used when the VM or GPU does not exist in the context of the request.
	// They are also used to overwrite the values derived from VM or GPU, which is useful when, for example,
	// moving resources between users.

	UserID         string
	Zone           *config.VmZone
	DeploymentZone *config.DeploymentZone
}
