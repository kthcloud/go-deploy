package v1_discover

import (
	"github.com/gin-gonic/gin"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service"
)

// Discover
// @Summary Discover
// @Description Discover
// @Tags Discover
// @Accept  json
// @Produce  json
// @Success 200 {object} body
// @Failure 500 {object} sys.ErrorResponse
// @Router /discover [get]
func Discover(c *gin.Context) {
	context := sys.NewContext(c)

	discover, err := service.V1().Discovery().Discover()
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	context.Ok(discover.ToDTO())
}
