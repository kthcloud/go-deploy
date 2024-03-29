package v1_sm

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-deploy/dto/v1/body"
	"go-deploy/dto/v1/query"
	"go-deploy/dto/v1/uri"
	"go-deploy/models/model"
	"go-deploy/models/versions"
	"go-deploy/pkg/app/status_codes"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service"
	"go-deploy/service/v1/sms/opts"
	v12 "go-deploy/service/v1/utils"
	"net/http"
)

// ListSMs
// @Summary Get storage manager list
// @Description Get storage manager list
// @BasePath /api/v1
// @Tags StorageManager
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {array} body.SmRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /storageManagers [get]storageManager
func ListSMs(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.SmList
	if err := context.GinContext.Bind(&requestQuery); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	smList, err := service.V1(auth).SMs().List(opts.ListOpts{
		Pagination: v12.GetOrDefaultPagination(requestQuery.Pagination),
		All:        requestQuery.All,
	})

	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if len(smList) == 0 {
		context.JSONResponse(200, []interface{}{})
		return
	}

	var smDTOs []body.SmRead
	for _, sm := range smList {
		smDTOs = append(smDTOs, sm.ToDTO())
	}

	context.Ok(smDTOs)
}

// GetSM
// @Summary Get storage manager
// @Description Get storage manager
// @BasePath /api/v1
// @Tags StorageManager
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param storageManagerId path string true "Storage manager ID"
// @Success 200 {object} body.SmDeleted
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /storageManagers/{storageManagerId} [get]
func GetSM(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.SmGet
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	sm, err := service.V1(auth).SMs().Get(requestURI.SmID)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.ResourceValidationFailed, "Failed to validate")
		return
	}

	if sm == nil || sm.OwnerID != auth.UserID {
		context.NotFound("Storage manager not found")
		return
	}

	context.Ok(sm.ToDTO())
}

// DeleteSM
// @Summary Delete storage manager
// @Description Delete storage manager
// @BasePath /api/v1
// @Tags StorageManager
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param storageManagerId path string true "Storage manager ID"
// @Success 200 {object} body.SmDeleted
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /storageManager/{storageManagerId} [get]
func DeleteSM(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.SmDelete
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	deployV1 := service.V1(auth)

	sm, err := deployV1.SMs().Get(requestURI.SmID)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if sm == nil {
		context.NotFound("Storage manager not found")
		return
	}

	jobID := uuid.New().String()
	err = deployV1.Jobs().Create(jobID, auth.UserID, model.JobDeleteSM, versions.V1, map[string]interface{}{
		"id": sm.ID,
	})
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	context.Ok(body.SmDeleted{
		ID:    sm.ID,
		JobID: jobID,
	})
}
