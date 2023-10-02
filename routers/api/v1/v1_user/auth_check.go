package v1_user

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/pkg/status_codes"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"net/http"
)

// AuthCheck
// @Summary Check auth
// @Description Check auth
// @Tags User
// @Accept  json
// @Produce  json
// @Success 200 {object} service.AuthInfo
// @Router /api/v1/user/auth-check [get]
func AuthCheck(c *gin.Context) {
	context := sys.NewContext(c)

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	c.JSON(200, auth)
}
