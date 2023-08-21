package v1_vm

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/pkg/status_codes"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"net/http"
)

func AccessGpuRoutes(c *gin.Context) {
	context := sys.NewContext(c)

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		c.Abort()
	}

	if !auth.GetEffectiveRole().Permissions.UseGPUs {
		context.ErrorResponse(http.StatusUnauthorized, status_codes.Error, "Tier does not include GPU access")
		c.Abort()
	}

	c.Next()
}
