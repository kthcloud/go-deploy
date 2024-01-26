package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-deploy/models/sys/event"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service"
	"go-deploy/utils"
)

// UserHttpEvent is a middleware that creates an event for the user's http request.
// This ensures that every request is logged in the database.
func UserHttpEvent() gin.HandlerFunc {
	return func(c *gin.Context) {
		context := sys.NewContext(c)

		auth, err := v1.WithAuth(&context)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to get auth when generating user http event. details: %w", err))
			return
		}

		userID := auth.UserID
		ip := context.GinContext.ClientIP()

		err = service.V1(auth).Events().Create(uuid.NewString(), &event.CreateParams{
			Type: event.TypeHttpRequest,
			Source: &event.Source{
				IP:     &ip,
				UserID: &userID,
			},
			Metadata: map[string]interface{}{
				"method": context.GinContext.Request.Method,
				"path":   context.GinContext.Request.URL.Path,
				"code":   context.GinContext.Writer.Status(),
			},
		})
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to create user http event. details: %w", err))
			return
		}

		c.Next()
	}
}
