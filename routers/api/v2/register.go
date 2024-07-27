package v2

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go-deploy/dto/v2/body"
	"go-deploy/pkg/sys"
	"go-deploy/service"
	sErrors "go-deploy/service/errors"
)

// Register
// @Summary Register resource
// @Description Register resource
// @Tags Register
// @Accept  json
// @Produce  json
// @Success 204
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/register [get]
func Register(c *gin.Context) {
	context := sys.NewContext(c)

	// Try parse body as body.HostRegisterParams
	var requestQueryJoin body.HostRegisterParams
	if err := context.GinContext.ShouldBindBodyWith(&requestQueryJoin, binding.JSON); err == nil {
		err = service.V2().System().RegisterNode(&requestQueryJoin)
		if err != nil {
			switch {
			case errors.Is(err, sErrors.BadDiscoveryTokenErr):
				context.UserError("Invalid token")
				return
			case errors.Is(err, sErrors.ZoneNotFoundErr):
				context.NotFound("Zone not found")
				return
			case errors.Is(err, sErrors.HostNotFoundErr):
				context.NotFound("Host not found")
				return
			}

			context.ServerError(err, fmt.Errorf("failed to register node"))
			return
		}

		context.OkNoContent()
		return
	}

	context.UserError("Invalid request body")
}
