package service

import (
	"github.com/kthcloud/go-deploy/service/clients"
	"github.com/kthcloud/go-deploy/service/core"
	"github.com/kthcloud/go-deploy/service/v2"
)

func V2(authInfo ...*core.AuthInfo) clients.V2 {
	return v2.New(authInfo...)
}
