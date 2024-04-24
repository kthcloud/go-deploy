package v1

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-deploy/dto/v1/body"
	"go-deploy/dto/v1/query"
	"go-deploy/dto/v1/uri"
	"go-deploy/models/model"
	"go-deploy/models/version"
	"go-deploy/pkg/sys"
	"go-deploy/service"
	"net/http"
)

// GetResourceMigration
// @Summary Get resource migration
// @Description Get resource migration
// @Tags ResourceMigration
// @Accept  json
// @Produce  json
// @Param resourceMigrationId path string true "Resource Migration ID"
// @Success 200 {object} body.ResourceMigrationRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v1/resourceMigrations/{resourceMigrationId} [get]
func GetResourceMigration(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery uri.ResourceMigrationGet
	if err := context.GinContext.ShouldBindUri(&requestQuery); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, AuthInfoNotAvailableErr)
		return
	}

	resourceMigration, err := service.V1(auth).ResourceMigrations().Get(requestQuery.ResourceMigrationID)
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	context.JSONResponse(http.StatusOK, resourceMigration.ToDTO())
}

// ListResourceMigrations
// @Summary List resource migrations
// @Description List resource migrations
// @Tags ResourceMigration
// @Accept  json
// @Produce  json
// @Success 200 {object} body.ResourceMigrationList
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v1/resourceMigrations [get]
func ListResourceMigrations(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.ResourceMigrationList
	if err := context.GinContext.ShouldBindQuery(&requestQuery); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, AuthInfoNotAvailableErr)
		return
	}

	resourceMigrations, err := service.V1(auth).ResourceMigrations().List()
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	if len(resourceMigrations) == 0 {
		context.JSONResponse(http.StatusOK, []interface{}{})
		return
	}

	dtoResourceMigrations := make([]interface{}, len(resourceMigrations))
	for i, resourceMigration := range resourceMigrations {
		dtoResourceMigrations[i] = resourceMigration.ToDTO()
	}

	context.JSONResponse(http.StatusOK, dtoResourceMigrations)
}

// CreateResourceMigration
// @Summary Create resource migration
// @Description Create resource migration
// @Tags ResourceMigration
// @Accept  json
// @Produce  json
// @Param body body body.ResourceMigrationCreate true "Resource Migration Create"
// @Success 201 {object} body.ResourceMigrationRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v1/resourceMigrations [post]
func CreateResourceMigration(c *gin.Context) {
	context := sys.NewContext(c)

	var requestBody body.ResourceMigrationCreate
	if err := context.GinContext.ShouldBindJSON(&requestBody); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, AuthInfoNotAvailableErr)
		return
	}

	deployV1 := service.V1(auth)

	resourceMigrationID := uuid.New().String()
	jobID := uuid.New().String()
	err = deployV1.Jobs().Create(jobID, auth.UserID, model.JobCreateResourceMigration, version.V1, map[string]interface{}{
		"id":     resourceMigrationID,
		"userId": auth.UserID,
		"params": requestBody,
	})

	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	context.Ok(body.ResourceMigrationCreated{
		ID:    resourceMigrationID,
		JobID: jobID,
	})
}

// UpdateResourceMigration
// @Summary Update resource migration
// @Description Update resource migration
// @Tags ResourceMigration
// @Accept  json
// @Produce  json
// @Param resourceMigrationId path string true "Resource Migration ID"
// @Param body body body.ResourceMigrationUpdate true "Resource Migration Update"
// @Success 200 {object} body.ResourceMigrationRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v1/resourceMigrations/{resourceMigrationId} [post]
func UpdateResourceMigration(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.ResourceMigrationUpdate
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	var requestQuery body.ResourceMigrationUpdate
	if err := context.GinContext.ShouldBindJSON(&requestQuery); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	var requestBody body.ResourceMigrationUpdate
	if err := context.GinContext.ShouldBindJSON(&requestBody); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, AuthInfoNotAvailableErr)
		return
	}

	deployV1 := service.V1(auth)

	resourceMigration, err := deployV1.ResourceMigrations().Get(requestURI.ResourceMigrationID)
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	if resourceMigration == nil {
		context.NotFound("Resource migration not found")
		return
	}

	jobID := uuid.New().String()
	err = deployV1.Jobs().Create(jobID, auth.UserID, model.JobUpdateResourceMigration, version.V1, map[string]interface{}{
		"id":     resourceMigration.ID,
		"params": requestBody,
	})

	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	context.Ok(body.ResourceMigrationUpdated{
		ID:    resourceMigration.ID,
		JobID: jobID,
	})
}

// DeleteResourceMigration
// @Summary Delete resource migration
// @Description Delete resource migration
// @Tags ResourceMigration
// @Accept  json
// @Produce  json
// @Param resourceMigrationId path string true "Resource Migration ID"
// @Success 204 {object} sys.EmptyResponse
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v1/resourceMigrations/{resourceMigrationId} [delete]
func DeleteResourceMigration(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery uri.ResourceMigrationDelete
	if err := context.GinContext.ShouldBindUri(&requestQuery); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, AuthInfoNotAvailableErr)
		return
	}

	deployV1 := service.V1(auth)

	resourceMigration, err := deployV1.ResourceMigrations().Get(requestQuery.ResourceMigrationID)
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	if resourceMigration == nil {
		context.NotFound("Resource migration not found")
		return
	}

	jobID := uuid.New().String()
	err = deployV1.Jobs().Create(jobID, auth.UserID, model.JobDeleteResourceMigration, version.V1, map[string]interface{}{
		"id": resourceMigration.ID,
	})

	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	context.Ok(body.ResourceMigrationDeleted{
		ID:    resourceMigration.ID,
		JobID: jobID,
	})
}
