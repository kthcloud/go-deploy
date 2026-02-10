package v2

import (
	"errors"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kthcloud/go-deploy/dto/v2/body"
	"github.com/kthcloud/go-deploy/dto/v2/query"
	"github.com/kthcloud/go-deploy/dto/v2/uri"
	configModels "github.com/kthcloud/go-deploy/models/config"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/models/version"
	"github.com/kthcloud/go-deploy/pkg/config"
	"github.com/kthcloud/go-deploy/pkg/sys"
	"github.com/kthcloud/go-deploy/service"
	sErrors "github.com/kthcloud/go-deploy/service/errors"
	teamOpts "github.com/kthcloud/go-deploy/service/v2/teams/opts"
	v2Utils "github.com/kthcloud/go-deploy/service/v2/utils"
	"github.com/kthcloud/go-deploy/service/v2/vms/opts"
)

// GetVM
// @Summary Get VM
// @Description Get VM
// @Tags VM
// @Produce  json
// @Security ApiKeyAuth
// @Security KeycloakOAuth
// @Param vmId path string true "VM ID"
// @Success 200 {object} body.VmRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/vms/{vmId} [get]
func GetVM(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.VmGet
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	var requestQuery query.VmGet
	if err := context.GinContext.ShouldBind(&requestQuery); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, ErrAuthInfoNotAvailable)
		return
	}

	deployV2 := service.V2(auth)

	var vm *model.VM
	if requestQuery.MigrationCode != nil {
		vm, err = deployV2.VMs().Get(requestURI.VmID, opts.GetOpts{MigrationCode: requestQuery.MigrationCode})
	} else {
		vm, err = deployV2.VMs().Get(requestURI.VmID, opts.GetOpts{Shared: true})
	}

	if err != nil {
		context.ServerError(err, ErrInternal)
		return
	}

	if vm == nil {
		context.NotFound("VM not found")
		return
	}

	teamIDs, _ := deployV2.Teams().ListIDs(teamOpts.ListOpts{ResourceID: vm.ID})
	sshConnectionString, _ := deployV2.VMs().SshConnectionString(vm.ID)

	lease, _ := deployV2.VMs().GpuLeases().GetByVmID(vm.ID)
	context.Ok(vm.ToDTOv2(lease, teamIDs, getVmAppExternalPort(vm.Zone), sshConnectionString))
}

// ListVMs
// @Summary List VMs
// @Description List VMs
// @Tags VM
// @Produce  json
// @Security ApiKeyAuth
// @Security KeycloakOAuth
// @Param all query bool false "List all"
// @Param userId query string false "Filter by user ID"
// @Param page query int false "Page number"
// @Param pageSize query int false "Number of items per page"
// @Success 200 {array} body.VmRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/vms [get]
func ListVMs(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.VmList
	if err := context.GinContext.ShouldBind(&requestQuery); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, ErrAuthInfoNotAvailable)
		return
	}

	deployV2 := service.V2(auth)

	var userID string
	if requestQuery.UserID != nil {
		userID = *requestQuery.UserID
	} else if !requestQuery.All {
		userID = auth.User.ID
	}

	vms, err := deployV2.VMs().List(opts.ListOpts{
		Pagination: v2Utils.GetOrDefaultPagination(requestQuery.Pagination),
		UserID:     &userID,
		Shared:     true,
	})
	if err != nil {
		context.ServerError(err, ErrInternal)
		return
	}

	if vms == nil {
		context.Ok([]interface{}{})
		return
	}

	dtoVMs := make([]body.VmRead, len(vms))
	for i, vm := range vms {
		teamIDs, _ := deployV2.Teams().ListIDs(teamOpts.ListOpts{ResourceID: vm.ID})
		sshConnectionString, _ := deployV2.VMs().SshConnectionString(vm.ID)
		lease, _ := deployV2.VMs().GpuLeases().GetByVmID(vm.ID)
		dtoVMs[i] = vm.ToDTOv2(lease, teamIDs, getVmAppExternalPort(vm.Zone), sshConnectionString)
	}

	context.Ok(dtoVMs)
}

// CreateVM
// @Summary Create VM
// @Description Create VM
// @Tags VM
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Security KeycloakOAuth
// @Param body body body.VmCreate true "VM body"
// @Success 200 {object} body.VmCreated
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 403 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/vms [post]
func CreateVM(c *gin.Context) {
	context := sys.NewContext(c)

	var requestBody body.VmCreate
	if err := context.GinContext.ShouldBindJSON(&requestBody); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, ErrAuthInfoNotAvailable)
		return
	}

	// Just to be sure (redundant)
	if auth != nil {
		// If the users role has explicitly specified false for useVms then we dont allow it.
		// To keep backward compatibility we allow roles without useVms specified to create vms.
		if role := auth.GetEffectiveRole(); role.Permissions.UseVms != nil && !*role.Permissions.UseVms {
			context.Forbidden("Not permitted to create a VM, missing useVms permission")
			return
		}
	}

	deployV2 := service.V2(auth)

	unique, err := deployV2.VMs().NameAvailable(requestBody.Name)
	if err != nil {
		context.ServerError(err, ErrInternal)
		return
	}

	if !unique {
		context.UserError("VM already exists")
		return
	}

	if requestBody.Zone != nil {
		zone := deployV2.System().GetZone(*requestBody.Zone)
		if zone == nil {
			context.NotFound("Zone not found")
			return
		}

		if !zone.Enabled {
			context.Forbidden("Zone is disabled")
			return
		}

		if !deployV2.System().ZoneHasCapability(*requestBody.Zone, configModels.ZoneCapabilityVM) {
			context.Forbidden("Zone does not have VM capability")
			return
		}
	}

	if requestBody.NeverStale && !auth.User.IsAdmin {
		context.Forbidden("User is not allowed to create VM with neverStale attribute set as true")
		return
	}

	err = deployV2.VMs().CheckQuota("", auth.User.ID, &auth.GetEffectiveRole().Quotas, opts.QuotaOpts{Create: &requestBody})
	if err != nil {
		var quotaExceedErr sErrors.QuotaExceededError
		if errors.As(err, &quotaExceedErr) {
			context.Forbidden(quotaExceedErr.Error())
			return
		}

		context.ServerError(err, ErrInternal)
		return
	}

	vmID := uuid.New().String()
	jobID := uuid.New().String()
	err = deployV2.Jobs().Create(jobID, auth.User.ID, model.JobCreateVM, version.V2, map[string]interface{}{
		"id":       vmID,
		"ownerId":  auth.User.ID,
		"params":   requestBody,
		"authInfo": auth,
	})
	if err != nil {
		context.ServerError(err, ErrInternal)
		return
	}

	context.Ok(body.VmCreated{
		ID:    vmID,
		JobID: jobID,
	})
}

// DeleteVM
// @Summary Delete VM
// @Description Delete VM
// @Tags VM
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Security KeycloakOAuth
// @Param vmId path string true "VM ID"
// @Success 200 {object} body.VmDeleted
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/vms/{vmId} [delete]
func DeleteVM(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.VmDelete
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, ErrAuthInfoNotAvailable)
		return
	}

	deployV2 := service.V2(auth)

	vm, err := deployV2.VMs().Get(requestURI.VmID, opts.GetOpts{Shared: true})
	if err != nil {
		context.ServerError(err, ErrInternal)
		return
	}

	if vm == nil {
		context.NotFound("VM not found")
		return
	}

	if vm.OwnerID != auth.User.ID && !auth.User.IsAdmin {
		context.Forbidden("VMs can only be deleted by their owner")
		return
	}

	jobID := uuid.New().String()
	err = deployV2.Jobs().Create(jobID, auth.User.ID, model.JobDeleteVM, version.V2, map[string]interface{}{
		"id":       vm.ID,
		"authInfo": auth,
	})
	if err != nil {
		context.ServerError(err, ErrInternal)
		return
	}

	context.Ok(body.VmDeleted{
		ID:    vm.ID,
		JobID: jobID,
	})
}

// UpdateVM
// @Summary Update VM
// @Description Update VM
// @Tags VM
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Security KeycloakOAuth
// @Param vmId path string true "VM ID"
// @Param body body body.VmUpdate true "VM update"
// @Success 200 {object} body.VmUpdated
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 403 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/vms/{vmId} [post]
func UpdateVM(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.VmUpdate
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	var requestBody body.VmUpdate
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

	vm, err := deployV2.VMs().Get(requestURI.VmID, opts.GetOpts{Shared: true})
	if err != nil {
		context.ServerError(err, ErrInternal)
		return
	}

	if vm == nil {
		context.NotFound("VM not found")
		return
	}

	if requestBody.Name != nil {
		available, err := deployV2.VMs().NameAvailable(*requestBody.Name)
		if err != nil {
			context.ServerError(err, ErrInternal)
			return
		}

		if !available {
			context.UserError("Name already taken")
			return
		}
	}

	if requestBody.Ports != nil {
		for _, port := range *requestBody.Ports {
			if port.HttpProxy != nil {
				// TODO: Fix this
				//available, err := deployV2.VMs().HttpProxyNameAvailable(requestURI.VmID, port.HttpProxy.Name)
				//if err != nil {
				//	context.ServerError(err, InternalError)
				//	return
				//}
				//
				//if !available {
				//	context.UserError("Http proxy name already taken")
				//	return
				//}
			}
		}
	}

	if requestBody.NeverStale != nil && !auth.User.IsAdmin {
		context.Forbidden("User is not allowed to modify the neverStale value")
		return
	}

	err = deployV2.VMs().CheckQuota(vm.ID, auth.User.ID, &auth.GetEffectiveRole().Quotas, opts.QuotaOpts{Update: &requestBody})
	if err != nil {
		var quotaExceededErr sErrors.QuotaExceededError
		if errors.As(err, &quotaExceededErr) {
			context.Forbidden(quotaExceededErr.Error())
			return
		}

		context.ServerError(err, ErrInternal)
		return
	}

	jobID := uuid.New().String()
	err = deployV2.Jobs().Create(jobID, auth.User.ID, model.JobUpdateVM, version.V2, map[string]interface{}{
		"id":       vm.ID,
		"params":   requestBody,
		"authInfo": auth,
	})
	if err != nil {
		context.ServerError(err, ErrInternal)
		return
	}

	context.Ok(body.VmUpdated{
		ID:    vm.ID,
		JobID: &jobID,
	})
}

func getVmAppExternalPort(zoneName string) *int {
	zone := config.Config.GetZone(zoneName)
	if zone == nil {
		return nil
	}

	split := strings.Split(zone.Domains.ParentVmApp, ":")
	if len(split) > 1 {
		port, err := strconv.Atoi(split[1])
		if err != nil {
			return nil
		}

		return &port
	}

	return nil
}
