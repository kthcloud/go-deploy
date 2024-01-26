package middleware

import (
	"github.com/gin-gonic/gin"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service"
)

// SynchronizeUser is a middleware that synchronizes the user's information in the database.
// This includes name, roles, permissions, etc.
func SynchronizeUser(c *gin.Context) {
	context := sys.NewContext(c)

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	_, err = service.V1(auth).Users().Create()
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	c.Next()
}
