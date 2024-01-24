package clients

import (
	"go-deploy/service/core"
	apiV2 "go-deploy/service/v2/api"
)

type V2 interface {
	Auth() *core.AuthInfo
	HasAuth() bool

	VMs() apiV2.VMs
}
