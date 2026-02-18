package v2

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kthcloud/go-deploy/dto/v2/body"
	"github.com/kthcloud/go-deploy/dto/v2/query"
	"github.com/kthcloud/go-deploy/dto/v2/uri"
	"github.com/kthcloud/go-deploy/pkg/sys"
	"github.com/kthcloud/go-deploy/service"
	sErrors "github.com/kthcloud/go-deploy/service/errors"
	"github.com/kthcloud/go-deploy/service/v2/resource_migrations/opts"
)

// GetResourceMigration
// @Summary Get resource migration
// @Description Get resource migration
// @Tags ResourceMigration
// @Produce json
// @Security ApiKeyAuth
// @Security KeycloakOAuth
// @Param resourceMigrationId path string true "Resource Migration ID"
// @Success 200 {object} body.ResourceMigrationRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/resourceMigrations/{resourceMigrationId} [get]
func GetResourceMigration(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery uri.ResourceMigrationGet
	if err := context.GinContext.ShouldBindUri(&requestQuery); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, ErrAuthInfoNotAvailable)
		return
	}

	resourceMigration, err := service.V2(auth).ResourceMigrations().Get(requestQuery.ResourceMigrationID)
	if err != nil {
		context.ServerError(err, ErrInternal)
		return
	}

	if resourceMigration == nil {
		context.NotFound("Resource migration not found")
		return
	}

	context.JSONResponse(http.StatusOK, resourceMigration.ToDTO())
}

// ListResourceMigrations
// @Summary List resource migrations
// @Description List resource migrations
// @Tags ResourceMigration
// @Produce json
// @Security ApiKeyAuth
// @Security KeycloakOAuth
// @Param page query int false "Page number"
// @Param pageSize query int false "Number of items per page"
// @Success 200 {array} body.ResourceMigrationRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/resourceMigrations [get]
func ListResourceMigrations(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.ResourceMigrationList
	if err := context.GinContext.ShouldBindQuery(&requestQuery); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, ErrAuthInfoNotAvailable)
		return
	}

	resourceMigrations, err := service.V2(auth).ResourceMigrations().List()
	if err != nil {
		context.ServerError(err, ErrInternal)
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
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Security KeycloakOAuth
// @Param body body body.ResourceMigrationCreate true "Resource Migration Create"
// @Success 200 {object} body.ResourceMigrationCreated
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/resourceMigrations [post]
func CreateResourceMigration(c *gin.Context) {
	context := sys.NewContext(c)

	var requestBody body.ResourceMigrationCreate
	if err := context.GinContext.ShouldBindJSON(&requestBody); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, ErrAuthInfoNotAvailable)
		return
	}

	deployV2 := service.V2(auth)

	resourceMigration, jobID, err := deployV2.ResourceMigrations().Create(uuid.New().String(), auth.User.ID, &requestBody)
	if err != nil {
		switch {
		case errors.Is(err, sErrors.ErrResourceMigrationAlreadyExists):
			context.UserError("Resource migration already exists")
			return
		case errors.Is(err, sErrors.ErrBadResourceMigrationParams):
			context.UserError("Bad resource migration params")
			return
		case errors.Is(err, sErrors.ErrBadResourceMigrationStatus):
			context.UserError("Bad resource migration status")
			return
		case errors.Is(err, sErrors.ErrBadResourceMigrationType):
			context.UserError("Bad resource migration type")
			return
		case errors.Is(err, sErrors.ErrBadResourceMigrationResourceType):
			context.UserError("Bad resource migration resource type")
			return
		case errors.Is(err, sErrors.ErrResourceNotFound):
			context.UserError("Resource not found")
			return
		case errors.Is(err, sErrors.ErrAlreadyMigrated):
			context.UserError("Resource already migrated")
			return
		}

		context.ServerError(err, ErrInternal)
		return
	}

	context.Ok(body.ResourceMigrationCreated{
		ResourceMigrationRead: resourceMigration.ToDTO(),
		JobID:                 jobID,
	})
}

// UpdateResourceMigration
// @Summary Update resource migration
// @Description Update resource migration
// @Tags ResourceMigration
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Security KeycloakOAuth
// @Param resourceMigrationId path string true "Resource Migration ID"
// @Param body body body.ResourceMigrationUpdate true "Resource Migration Update"
// @Success 200 {object} body.ResourceMigrationUpdated
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/resourceMigrations/{resourceMigrationId} [post]
func UpdateResourceMigration(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.ResourceMigrationUpdate
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
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
		context.ServerError(err, ErrAuthInfoNotAvailable)
		return
	}

	deployV2 := service.V2(auth)

	resourceMigration, err := deployV2.ResourceMigrations().Get(requestURI.ResourceMigrationID, opts.GetOpts{
		MigrationCode: requestBody.Code,
	})
	if err != nil {
		context.ServerError(err, ErrInternal)
		return
	}

	if resourceMigration == nil {
		context.NotFound("Resource migration not found")
		return
	}

	var jobID *string
	resourceMigration, jobID, err = deployV2.ResourceMigrations().Update(resourceMigration.ID, &requestBody, opts.UpdateOpts{
		MigrationCode: requestBody.Code,
	})
	if err != nil {
		switch {
		case errors.Is(err, sErrors.ErrResourceMigrationNotFound):
			context.NotFound("Resource migration not found")
			return
		case errors.Is(err, sErrors.ErrAlreadyAccepted):
			context.UserError("Resource migration already accepted")
			return
		case errors.Is(err, sErrors.ErrAlreadyMigrated):
			context.UserError("Resource already migrated")
			return
		case errors.Is(err, sErrors.ErrBadResourceMigrationParams):
			context.UserError("Bad resource migration params")
			return
		case errors.Is(err, sErrors.ErrBadResourceMigrationStatus):
			context.UserError("Bad resource migration status")
			return
		case errors.Is(err, sErrors.ErrBadResourceMigrationType):
			context.UserError("Bad resource migration type")
			return
		case errors.Is(err, sErrors.ErrBadResourceMigrationResourceType):
			context.UserError("Bad resource migration resource type")
			return
		case errors.Is(err, sErrors.ErrResourceNotFound):
			context.UserError("Resource not found")
			return
		case errors.Is(err, sErrors.ErrBadMigrationCode):
			context.UserError("Bad migration code")
			return
		}

		context.ServerError(err, ErrInternal)
		return
	}

	if resourceMigration == nil {
		context.NotFound("Resource migration not found")
		return
	}

	context.Ok(body.ResourceMigrationUpdated{
		ResourceMigrationRead: resourceMigration.ToDTO(),
		JobID:                 jobID,
	})
}

// DeleteResourceMigration
// @Summary Delete resource migration
// @Description Delete resource migration
// @Tags ResourceMigration
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Security KeycloakOAuth
// @Param resourceMigrationId path string true "Resource Migration ID"
// @Success 204 "No Content"
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/resourceMigrations/{resourceMigrationId} [delete]
func DeleteResourceMigration(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery uri.ResourceMigrationDelete
	if err := context.GinContext.ShouldBindUri(&requestQuery); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, ErrAuthInfoNotAvailable)
		return
	}

	deployV2 := service.V2(auth)

	resourceMigration, err := deployV2.ResourceMigrations().Get(requestQuery.ResourceMigrationID)
	if err != nil {
		context.ServerError(err, ErrInternal)
		return
	}

	if resourceMigration == nil {
		context.NotFound("Resource migration not found")
		return
	}

	err = deployV2.ResourceMigrations().Delete(resourceMigration.ID)
	if err != nil {
		context.ServerError(err, ErrInternal)
		return
	}

	context.OkNoContent()
}
