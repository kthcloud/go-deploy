package service

import (
	"go-deploy/service/clients"
	"go-deploy/service/core"
	"go-deploy/service/v1"
	"go-deploy/service/v2"
)

func V1(authInfo ...*core.AuthInfo) clients.V1 {
	return v1.New(authInfo...)
}

func V2(authInfo ...*core.AuthInfo) clients.V2 {
	return v2.New(v1.New(authInfo...), authInfo...)
}
