package service

import (
	"go-deploy/service/clients"
	"go-deploy/service/core"
	"go-deploy/service/v2"
)

func V2(authInfo ...*core.AuthInfo) clients.V2 {
	return v2.New(authInfo...)
}
