package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-deploy/models/model"
	"go-deploy/models/version"
	"go-deploy/pkg/config"
	"go-deploy/pkg/sys"
	v2 "go-deploy/routers/api/v2"
	"go-deploy/service"
	"go-deploy/utils"
)

// CreateSM is a middleware that creates a storage manager for the user if it does not exist.
// The storage manager is created asynchronously, so it may not be available immediately.
func CreateSM() gin.HandlerFunc {
	return func(c *gin.Context) {
		context := sys.NewContext(c)

		auth, err := v2.WithAuth(&context)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to get auth when creating storage manager. details: %w", err))
			return
		}

		deployV2 := service.V2(auth)

		exists, err := deployV2.SMs().Exists(auth.User.ID)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to create storage manager. details: %w", err))
			return
		}

		if !exists {
			smID := uuid.New().String()
			jobID := uuid.New().String()
			err = deployV2.Jobs().Create(jobID, auth.User.ID, model.JobCreateSM, version.V2, map[string]interface{}{
				"id":     smID,
				"userId": auth.User.ID,
				"params": model.SmCreateParams{
					Zone: config.Config.Deployment.DefaultZone,
				},
				"authInfo": auth,
			})

			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("failed to create storage manager (if not exists). details: %w", err))
				return
			}
		}

		c.Next()
	}
}
