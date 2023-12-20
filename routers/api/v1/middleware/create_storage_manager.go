package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	jobModel "go-deploy/models/sys/job"
	"go-deploy/models/sys/sm"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/job_service"
	"go-deploy/service/sm_service"
	"go-deploy/utils"
)

func CreateSM() gin.HandlerFunc {
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
			smID := uuid.New().String()
			jobID := uuid.New().String()
			err = job_service.New().Create(jobID, auth.UserID, jobModel.TypeCreateSM, map[string]interface{}{
				"id":     smID,
				"userId": auth.UserID,
				"params": sm.CreateParams{
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
