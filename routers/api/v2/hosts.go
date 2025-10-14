package v2

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/kthcloud/go-deploy/dto/v2/body"
	"github.com/kthcloud/go-deploy/pkg/sys"
	"github.com/kthcloud/go-deploy/service"
)

// ListHosts
// @Summary List Hosts
// @Description List Hosts
// @Tags Host
// @Produce  json
// @Success 200 {array} body.HostRead
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/hosts [get]
func ListHosts(c *gin.Context) {
	context := sys.NewContext(c)

	hostInfo, err := service.V2().System().ListHosts()
	if err != nil {
		context.ServerError(err, fmt.Errorf("failed to get host info"))
	}
	dtoHosts := make([]body.HostRead, 0)
	for _, host := range hostInfo {
		dtoHosts = append(dtoHosts, host.ToDTO())
	}
	context.JSONResponse(200, dtoHosts)
}

// VerboseListHosts
// @Summary List Hosts verbose
// @Description List Hosts verbose
// @Tags Host
// @Produce  json
// @Security ApiKeyAuth
// @Security KeycloakOAuth
// @Success 200 {array} body.HostVerboseRead
// @Failure 403 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/hosts/verbose [get]
func VerboseListHosts(c *gin.Context) {
	context := sys.NewContext(c)

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, ErrAuthInfoNotAvailable)
		return
	}

	if (auth != nil && !auth.User.IsAdmin) || auth == nil {
		context.Forbidden("Not permitted to request verbose host information")
		return
	}

	hostInfo, err := service.V2().System().ListAllHosts()
	if err != nil {
		context.ServerError(err, fmt.Errorf("failed to get host info"))
	}
	dtoHosts := make([]body.HostVerboseRead, 0)
	for _, host := range hostInfo {
		dtoHosts = append(dtoHosts, host.ToVerboseDTO())
	}
	context.JSONResponse(200, dtoHosts)
}
