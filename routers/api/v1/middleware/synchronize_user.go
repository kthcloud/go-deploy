package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/pkg/app/status_codes"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/user_service"
	"net/http"
)

func SynchronizeUser(c *gin.Context) {
	context := sys.NewContext(c)

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	_, err = user_service.Create(auth)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get user: %s", err))
		return
	}

	c.Next()
}
