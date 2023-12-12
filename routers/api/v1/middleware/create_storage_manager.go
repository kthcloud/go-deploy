package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	jobModel "go-deploy/models/sys/job"
	"go-deploy/models/sys/storage_manager"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/job_service"
	"go-deploy/service/sm_service"
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

		exists, err := sm_service.New().Exists(auth.UserID)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to create storage manager. details: %w", err))
			return
		}

		if !exists {
			storageManagerID := uuid.New().String()
			jobID := uuid.New().String()
			err = job_service.Create(jobID, auth.UserID, jobModel.TypeCreateStorageManager, map[string]interface{}{
				"id":     storageManagerID,
				"userId": auth.UserID,
				"params": storage_manager.CreateParams{
					Zone: "se-flem",
				},
			})

			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("failed to create storage manager (if not exists). details: %w", err))
				return
			}
		}

		c.Next()
	}
}
