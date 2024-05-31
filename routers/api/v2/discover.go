package v2

import (
	"github.com/gin-gonic/gin"
	"go-deploy/pkg/sys"
	"go-deploy/service"
)

// Discover
// @Summary Discover
// @Description Discover
// @Tags Discover
// @Accept  json
// @Produce  json
// @Success 200 {object} body.DiscoverRead
// @Failure 500 {object} sys.ErrorResponse
// @Router /v1/discover [get]
func Discover(c *gin.Context) {
	context := sys.NewContext(c)

	discover, err := service.V1().Discovery().Discover()
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	context.Ok(discover.ToDTO())
}
