package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/kthcloud/go-deploy/pkg/app/status_codes"
	"github.com/kthcloud/go-deploy/pkg/sys"
	v2 "github.com/kthcloud/go-deploy/routers/api/v2"
	"net/http"
)

// AccessGpuRoutes is a middleware that checks if the user has access to the GPU routes.
// If the user does not have access, the request is aborted with http.StatusForbidden.
func AccessGpuRoutes() func(c *gin.Context) {
	return func(c *gin.Context) {
		context := sys.NewContext(c)

		auth, err := v2.WithAuth(&context)
		if err != nil {
			context.ServerError(err, v2.AuthInfoNotAvailableErr)
			c.Abort()
		}

		if !auth.GetEffectiveRole().Permissions.UseGPUs && !auth.User.IsAdmin {
			context.ErrorResponse(http.StatusForbidden, status_codes.Error, "Tier does not include GPU access")
			c.Abort()
		}

		c.Next()
	}
}
