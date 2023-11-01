package v1_user

import (
	"github.com/gin-gonic/gin"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
)

// AuthCheck
// @Summary Check auth
// @Description Check auth
// @Tags User
// @Accept  json
// @Produce  json
// @Success 200 {object} service.AuthInfo
// @Router /user/auth-check [get]
func AuthCheck(c *gin.Context) {
	context := sys.NewContext(c)

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	c.JSON(200, auth)
}
