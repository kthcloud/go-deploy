package middleware

import (
	"github.com/gin-gonic/gin"
	"go-deploy/pkg/app/status_codes"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"net/http"
)

func AccessGpuRoutes() func(c *gin.Context) {
	return func(c *gin.Context) {
		context := sys.NewContext(c)

		auth, err := v1.WithAuth(&context)
		if err != nil {
			context.ServerError(err, v1.AuthInfoNotAvailableErr)
			c.Abort()
		}

		if !auth.GetEffectiveRole().Permissions.UseGPUs && !auth.IsAdmin {
			context.ErrorResponse(http.StatusForbidden, status_codes.Error, "Tier does not include GPU access")
			c.Abort()
		}

		c.Next()
	}
}
