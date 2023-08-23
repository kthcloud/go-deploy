package v1_deployment

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/query"
	"go-deploy/models/dto/uri"
	jobModel "go-deploy/models/sys/job"
	"go-deploy/pkg/status_codes"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/deployment_service"
	"go-deploy/service/job_service"
	"net/http"
)


// GetStorageManagerList
// @Summary Get storage manager list
// @Description Get storage manager list
// @BasePath /api/v1
// @Tags Deployment
// @Accept json
// @Produce json
// @Param Authorization header string true "With the bearer started"
// @Success 200 {array} body.StorageManagerRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /api/v1/storageManagers [get]
func GetStorageManagerList(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.StorageManagerList
	if err := context.GinContext.Bind(&requestQuery); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err.Error()))
		return
	}

	if requestQuery.WantAll {
		storageManagers, _ := deployment_service.GetAllStorageManagers(auth)

		dtoStorageManagers := make([]body.StorageManagerRead, len(storageManagers))
		for i, deployment := range storageManagers {
			dtoStorageManagers[i] = deployment.ToDTO()
		}

		context.JSONResponse(http.StatusOK, dtoStorageManagers)
		return
	}

	storageManagers, _ := deployment_service.GetStorageManagerByOwnerID(auth)
	if storageManagers == nil {
		context.JSONResponse(200, []interface{}{})
		return
	}

	dtoDeployments := make([]body.StorageManagerRead, len(storageManagers))
	for i, storageManager := range storageManagers {
		dtoDeployments[i] = storageManager.ToDTO()
	}

	context.JSONResponse(200, dtoDeployments)
}

// GetStorageManager
// @Summary Get storage manager
// @Description Get storage manager
// @BasePath /api/v1
// @Tags Deployment
// @Accept json
// @Produce json
// @Param Authorization header string true "With the bearer started"
// @Param storageManagerId path string true "Storage manager ID"
// @Success 200 {object} body.StorageManagerDeleted
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /api/v1/storageManagers/{storageManagerId} [get]
func GetStorageManager(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.StorageManagerGet
	if err := context.GinContext.BindUri(&requestURI); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	storageManager, err := deployment_service.GetStorageManagerByID(requestURI.StorageManagerID, auth)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.ResourceValidationFailed, "Failed to validate")
		return
	}

	if storageManager == nil || storageManager.OwnerID != auth.UserID {
		context.NotFound()
		return
	}

	context.JSONResponse(http.StatusOK, storageManager.ToDTO())
}

// DeleteStorageManager
// @Summary Delete storage manager
// @Description Delete storage manager
// @BasePath /api/v1
// @Tags Deployment
// @Accept json
// @Produce json
// @Param Authorization header string true "With the bearer started"
// @Param storageManagerId path string true "Storage manager ID"
// @Success 200 {object} body.StorageManagerDeleted
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /api/v1/storageManager/{storageManagerId} [get]
func DeleteStorageManager(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.StorageManagerDelete
	if err := context.GinContext.BindUri(&requestURI); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	storageManager, err := deployment_service.GetStorageManagerByID(requestURI.StorageManagerID, auth)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.ResourceValidationFailed, "Failed to validate")
		return
	}

	if storageManager == nil || storageManager.OwnerID != auth.UserID {
		context.NotFound()
		return
	}

	jobID := uuid.New().String()
	err = job_service.Create(jobID, auth.UserID, jobModel.TypeDeleteStorageManager, map[string]interface{}{
		"id": storageManager.ID,
	})
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to create job: %s", err))
		return
	}

	context.JSONResponse(http.StatusOK, body.StorageManagerDeleted{
		ID:    storageManager.ID,
		JobID: jobID,
	})
}
