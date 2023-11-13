package v1_deployment

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/query"
	"go-deploy/models/dto/uri"
	jobModel "go-deploy/models/sys/job"
	"go-deploy/pkg/app/status_codes"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/job_service"
	"go-deploy/service/storage_manager_service"
	"net/http"
)

// ListStorageManagers
// @Summary Get storage manager list
// @Description Get storage manager list
// @BasePath /api/v1
// @Tags Deployment
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {array} body.StorageManagerRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /storageManagers [get]storageManager
func ListStorageManagers(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.StorageManagerList
	if err := context.GinContext.Bind(&requestQuery); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	if requestQuery.All {
		storageManagers, _ := storage_manager_service.GetAll(auth)

		dtoStorageManagers := make([]body.StorageManagerRead, len(storageManagers))
		for i, deployment := range storageManagers {
			dtoStorageManagers[i] = deployment.ToDTO()
		}

		context.JSONResponse(http.StatusOK, dtoStorageManagers)
		return
	}

	storageManagers, err := storage_manager_service.ListAuth(requestQuery.All, requestQuery.UserID, auth, &requestQuery.Pagination)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if len(storageManagers) == 0 {
		context.JSONResponse(200, []interface{}{})
		return
	}

	var storageManagerDTOs []body.StorageManagerRead
	for _, storageManager := range storageManagers {
		storageManagerDTOs = append(storageManagerDTOs, storageManager.ToDTO())
	}

	context.Ok(storageManagerDTOs)
}

// GetStorageManager
// @Summary Get storage manager
// @Description Get storage manager
// @BasePath /api/v1
// @Tags Deployment
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param storageManagerId path string true "Storage manager ID"
// @Success 200 {object} body.StorageManagerDeleted
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /storageManagers/{storageManagerId} [get]
func GetStorageManager(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.StorageManagerGet
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	storageManager, err := storage_manager_service.GetByIdAuth(requestURI.StorageManagerID, auth)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.ResourceValidationFailed, "Failed to validate")
		return
	}

	if storageManager == nil || storageManager.OwnerID != auth.UserID {
		context.NotFound("Storage manager not found")
		return
	}

	context.Ok(storageManager.ToDTO())
}

// DeleteStorageManager
// @Summary Delete storage manager
// @Description Delete storage manager
// @BasePath /api/v1
// @Tags Deployment
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param storageManagerId path string true "Storage manager ID"
// @Success 200 {object} body.StorageManagerDeleted
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /storageManager/{storageManagerId} [get]
func DeleteStorageManager(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.StorageManagerDelete
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	storageManager, err := storage_manager_service.GetByIdAuth(requestURI.StorageManagerID, auth)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if storageManager == nil {
		context.NotFound("Storage manager not found")
		return
	}

	jobID := uuid.New().String()
	err = job_service.Create(jobID, auth.UserID, jobModel.TypeDeleteStorageManager, map[string]interface{}{
		"id": storageManager.ID,
	})
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	context.Ok(body.StorageManagerDeleted{
		ID:    storageManager.ID,
		JobID: jobID,
	})
}
