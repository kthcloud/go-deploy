package client

import (
	"go-deploy/models/sys/gpu"
	"go-deploy/models/sys/vm"
	"go-deploy/service"
)

type Context struct {
	vmStore  map[string]*vm.VM
	gpuStore map[string]*gpu.GPU

	Auth *service.AuthInfo
}
