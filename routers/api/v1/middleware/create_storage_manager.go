package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	jobModels "go-deploy/models/sys/job"
	"go-deploy/models/sys/sm"
	"go-deploy/models/versions"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service"
	"go-deploy/utils"
)

// CreateSM is a middleware that creates a storage manager for the user if it does not exist.
// The storage manger is created asynchronously, so it may not be available immediately.
func CreateSM() gin.HandlerFunc {
	return func(c *gin.Context) {
		context := sys.NewContext(c)

		auth, err := v1.WithAuth(&context)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to get auth when creating storage manager. details: %w", err))
			return
		}

		deployV1 := service.V1(auth)

		exists, err := deployV1.SMs().Exists(auth.UserID)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to create storage manager. details: %w", err))
			return
		}

		if !exists {
			smID := uuid.New().String()
			jobID := uuid.New().String()
			err = deployV1.Jobs().Create(jobID, auth.UserID, jobModels.TypeCreateSM, versions.V1, map[string]interface{}{
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
