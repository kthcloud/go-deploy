package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/storage_manager_service"
	"go-deploy/utils"
)

func CreateStorageManager() gin.HandlerFunc {
	return func(c *gin.Context) {
		context := sys.NewContext(c)

		auth, err := v1.WithAuth(&context)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to get auth when creating storage manager. details: %w", err))
			return
		}

		err = storage_manager_service.New().WithUserID(auth.UserID).CreateIfNotExists()
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to create storage manager. details: %w", err))
			return
		}

		c.Next()
	}
}
