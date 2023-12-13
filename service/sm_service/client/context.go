package client

import (
	smModels "go-deploy/models/sys/sm"
	"go-deploy/service"
)

type Context struct {
	smStore map[string]*smModels.SM

	Auth *service.AuthInfo
}
