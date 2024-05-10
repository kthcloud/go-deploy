package v2

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-deploy/dto/v2/body"
	"go-deploy/dto/v2/query"
	"go-deploy/dto/v2/uri"
	configModels "go-deploy/models/config"
	"go-deploy/models/model"
	"go-deploy/models/version"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service"
	sErrors "go-deploy/service/errors"
	teamOpts "go-deploy/service/v1/teams/opts"
	v2Utils "go-deploy/service/v2/utils"
	"go-deploy/service/v2/vms/opts"
)

// GetVM
// @Summary Get VM
// @Description Get VM
// @Tags VM
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
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
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	var requestQuery query.VmGet
	if err := context.GinContext.ShouldBind(&requestQuery); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	deployV1 := service.V1(auth)
	deployV2 := service.V2(auth)

	var vm *model.VM
	if requestQuery.MigrationCode != nil {
		vm, err = deployV2.VMs().Get(requestURI.VmID, opts.GetOpts{MigrationCode: requestQuery.MigrationCode})
	} else {
		vm, err = deployV2.VMs().Get(requestURI.VmID, opts.GetOpts{Shared: true})
	}

	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if vm == nil {
		context.NotFound("VM not found")
		return
	}

	teamIDs, _ := deployV1.Teams().ListIDs(teamOpts.ListOpts{ResourceID: vm.ID})
	sshConnectionString, _ := deployV2.VMs().SshConnectionString(vm.ID)

	lease, _ := deployV2.VMs().GpuLeases().GetByVmID(vm.ID)
	context.Ok(vm.ToDTOv2(lease, teamIDs, sshConnectionString))
}

// ListVMs
// @Summary List VMs
// @Description List VMs
// @Tags VM
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
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
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	deployV1 := service.V1(auth)
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
		context.ServerError(err, v1.InternalError)
		return
	}

	if vms == nil {
		context.Ok([]interface{}{})
		return
	}

	dtoVMs := make([]body.VmRead, len(vms))
	for i, vm := range vms {
		teamIDs, _ := deployV1.Teams().ListIDs(teamOpts.ListOpts{ResourceID: vm.ID})
		sshConnectionString, _ := deployV2.VMs().SshConnectionString(vm.ID)
		lease, _ := deployV2.VMs().GpuLeases().GetByVmID(vm.ID)
		dtoVMs[i] = vm.ToDTOv2(lease, teamIDs, sshConnectionString)
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
// @Param body body body.VmCreate true "VM body"
// @Success 200 {object} body.VmCreated
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/vms [post]
func CreateVM(c *gin.Context) {
	context := sys.NewContext(c)

	var requestBody body.VmCreate
	if err := context.GinContext.ShouldBindJSON(&requestBody); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	deployV1 := service.V1(auth)
	deployV2 := service.V2(auth)

	unique, err := deployV2.VMs().NameAvailable(requestBody.Name)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if !unique {
		context.UserError("VM already exists")
		return
	}

	if requestBody.Zone != nil {
		zone := deployV1.Zones().Get(*requestBody.Zone)
		if zone == nil {
			context.NotFound("Zone not found")
			return
		}

		if !zone.Enabled {
			context.Forbidden("Zone is disabled")
			return
		}

		if !deployV1.Zones().HasCapability(*requestBody.Zone, configModels.ZoneCapabilityVM) {
			context.Forbidden("Zone does not have VM capability")
			return
		}
	}

	err = deployV2.VMs().CheckQuota("", auth.User.ID, &auth.GetEffectiveRole().Quotas, opts.QuotaOpts{Create: &requestBody})
	if err != nil {
		var quotaExceedErr sErrors.QuotaExceededError
		if errors.As(err, &quotaExceedErr) {
			context.Forbidden(quotaExceedErr.Error())
			return
		}

		context.ServerError(err, v1.InternalError)
		return
	}

	vmID := uuid.New().String()
	jobID := uuid.New().String()
	err = deployV1.Jobs().Create(jobID, auth.User.ID, model.JobCreateVM, version.V2, map[string]interface{}{
		"id":       vmID,
		"ownerId":  auth.User.ID,
		"params":   requestBody,
		"authInfo": auth,
	})
	if err != nil {
		context.ServerError(err, v1.InternalError)
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
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	deployV1 := service.V1(auth)
	deployV2 := service.V2(auth)

	vm, err := deployV2.VMs().Get(requestURI.VmID, opts.GetOpts{Shared: true})
	if err != nil {
		context.ServerError(err, v1.InternalError)
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
	err = deployV1.Jobs().Create(jobID, auth.User.ID, model.JobDeleteVM, version.V2, map[string]interface{}{
		"id":       vm.ID,
		"authInfo": auth,
	})
	if err != nil {
		context.ServerError(err, v1.InternalError)
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
// @Param vmId path string true "VM ID"
// @Param body body body.VmUpdate true "VM update"
// @Success 200 {object} body.VmUpdated
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/vms/{vmId} [post]
func UpdateVM(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.VmUpdate
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	var requestBody body.VmUpdate
	if err := context.GinContext.ShouldBindJSON(&requestBody); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	deployV1 := service.V1(auth)
	deployV2 := service.V2(auth)

	vm, err := deployV2.VMs().Get(requestURI.VmID, opts.GetOpts{Shared: true})
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if vm == nil {
		context.NotFound("VM not found")
		return
	}

	if requestBody.Name != nil {
		available, err := deployV2.VMs().NameAvailable(*requestBody.Name)
		if err != nil {
			context.ServerError(err, v1.InternalError)
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
				//	context.ServerError(err, v1.InternalError)
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

	err = deployV2.VMs().CheckQuota(vm.ID, auth.User.ID, &auth.GetEffectiveRole().Quotas, opts.QuotaOpts{Update: &requestBody})
	if err != nil {
		var quotaExceededErr sErrors.QuotaExceededError
		if errors.As(err, &quotaExceededErr) {
			context.Forbidden(quotaExceededErr.Error())
			return
		}

		context.ServerError(err, v1.InternalError)
		return
	}

	jobID := uuid.New().String()
	err = deployV1.Jobs().Create(jobID, auth.User.ID, model.JobUpdateVM, version.V2, map[string]interface{}{
		"id":       vm.ID,
		"params":   requestBody,
		"authInfo": auth,
	})

	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	context.Ok(body.VmUpdated{
		ID:    vm.ID,
		JobID: &jobID,
	})
}
