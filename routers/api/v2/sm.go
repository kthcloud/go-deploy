package v2

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kthcloud/go-deploy/dto/v2/body"
	"github.com/kthcloud/go-deploy/dto/v2/query"
	"github.com/kthcloud/go-deploy/dto/v2/uri"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/models/version"
	"github.com/kthcloud/go-deploy/pkg/app/status_codes"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/sys"
	"github.com/kthcloud/go-deploy/service"
	"github.com/kthcloud/go-deploy/service/v2/sms/opts"
	v12 "github.com/kthcloud/go-deploy/service/v2/utils"
)

// GetSM
// @Summary Get storage manager
// @Description Get storage manager
// @Tags StorageManager
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Security KeycloakOAuth
// @Param storageManagerId path string true "Storage manager ID"
// @Success 200 {object} body.SmDeleted
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/storageManagers/{storageManagerId} [get]
func GetSM(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.SmGet
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, AuthInfoNotAvailableErr)
		return
	}

	sm, err := service.V2(auth).SMs().Get(requestURI.SmID)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.ResourceValidationFailed, "Failed to validate")
		return
	}

	if sm == nil {
		context.NotFound("Storage manager not found")
		return
	}

	context.Ok(sm.ToDTO(getSmExternalPort(sm.Zone)))
}

// ListSMs
// @Summary Get storage manager list
// @Description Get storage manager list
// @Tags StorageManager
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Security KeycloakOAuth
// @Param all query bool false "List all"
// @Param page query int false "Page number"
// @Param pageSize query int false "Number of items per page"
// @Success 200 {array} body.SmRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/storageManagers [get]storageManager
func ListSMs(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.SmList
	if err := context.GinContext.ShouldBind(&requestQuery); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, AuthInfoNotAvailableErr)
		return
	}

	smList, err := service.V2(auth).SMs().List(opts.ListOpts{
		Pagination: v12.GetOrDefaultPagination(requestQuery.Pagination),
		All:        requestQuery.All,
	})

	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	if len(smList) == 0 {
		context.JSONResponse(200, []interface{}{})
		return
	}

	var smDTOs []body.SmRead
	for _, sm := range smList {
		smDTOs = append(smDTOs, sm.ToDTO(getSmExternalPort(sm.Zone)))
	}

	context.Ok(smDTOs)
}

// DeleteSM
// @Summary Delete storage manager
// @Description Delete storage manager
// @Tags StorageManager
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Security KeycloakOAuth
// @Param storageManagerId path string true "Storage manager ID"
// @Success 200 {object} body.SmDeleted
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/storageManagers/{storageManagerId} [get]
func DeleteSM(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.SmDelete
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, AuthInfoNotAvailableErr)
		return
	}

	deployV2 := service.V2(auth)

	sm, err := deployV2.SMs().Get(requestURI.SmID)
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	if sm == nil {
		context.NotFound("Storage manager not found")
		return
	}

	jobID := uuid.New().String()
	err = deployV2.Jobs().Create(jobID, auth.User.ID, model.JobDeleteSM, version.V2, map[string]interface{}{
		"id":      sm.ID,
		"ownerId": sm.OwnerID,
	})
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	context.Ok(body.SmDeleted{
		ID:    sm.ID,
		JobID: jobID,
	})
}

func getSmExternalPort(zoneName string) *int {
	zone := config.Config.GetZone(zoneName)
	if zone == nil {
		return nil
	}

	split := strings.Split(zone.Domains.ParentSM, ":")
	if len(split) > 1 {
		port, err := strconv.Atoi(split[1])
		if err != nil {
			return nil
		}

		return &port
	}

	return nil
}
