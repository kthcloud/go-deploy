package v2

import (
	"github.com/gin-gonic/gin"
	"github.com/kthcloud/go-deploy/pkg/sys"
	"github.com/kthcloud/go-deploy/service"
)

// Discover
// @Summary Discover
// @Description Discover
// @Tags Discover
// @Produce  json
// @Success 200 {object} body.DiscoverRead
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/discover [get]
func Discover(c *gin.Context) {
	context := sys.NewContext(c)

	discover, err := service.V2().Discovery().Discover()
	if err != nil {
		context.ServerError(err, ErrInternal)
		return
	}

	context.Ok(discover.ToDTO())
}
