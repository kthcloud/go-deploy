package v2

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-deploy/dto/v2/body"
	"go-deploy/dto/v2/query"
	"go-deploy/dto/v2/uri"
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
// @Summary GetVM VM by id
// @Description GetVM VM by id
// @Tags VM
// @Accept  json
// @Produce  json
// @Param vmId path string true "VM ID"
// @Success 200 {object} body.VmRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /vm/{vmId} [get]
func GetVM(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.VmGet
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

	teamIDs, _ := deployV1.Teams().ListIDs(teamOpts.ListOpts{ResourceID: vm.ID})
	sshConnectionString, _ := deployV2.VMs().SshConnectionString(vm.ID)

	lease, _ := deployV2.VMs().GpuLeases().GetByVmID(vm.ID)
	context.Ok(vm.ToDTOv2(lease, teamIDs, sshConnectionString))
}

// ListVMs
// @Summary GetVM list of VMs
// @Description GetVM list of VMs
// @Tags VM
// @Accept  json
// @Produce  json
// @Param all query bool false "GetVM all"
// @Param userId query string false "Filter by user id"
// @Param page query int false "Page number"
// @Param pageSize query int false "Number of items per page"
// @Success 200 {array} body.VmRead
// @Failure 500 {object} sys.ErrorResponse
// @Router /vm [get]
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
		userID = auth.UserID
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
// @Summary CreateVM VM
// @Description CreateVM VM
// @Tags VM
// @Accept  json
// @Produce  json
// @Param body body body.VmCreate true "VM body"
// @Success 200 {object} body.VmCreated
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 423 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /vm [post]
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
		zone := deployV1.Zones().Get(*requestBody.Zone, model.ZoneTypeVM)
		if zone == nil {
			context.NotFound("Zone not found")
			return
		}
	}

	err = deployV2.VMs().CheckQuota("", auth.UserID, &auth.GetEffectiveRole().Quotas, opts.QuotaOpts{Create: &requestBody})
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
	err = deployV1.Jobs().Create(jobID, auth.UserID, model.JobCreateVM, version.V2, map[string]interface{}{
		"id":      vmID,
		"ownerId": auth.UserID,
		"params":  requestBody,
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
// @Param vmId path string true "VM ID"
// @Success 200 {object} body.VmDeleted
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 423 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /vm/{vmId} [delete]
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

	if vm.OwnerID != auth.UserID && !auth.IsAdmin {
		context.Forbidden("VMs can only be deleted by their owner")
		return
	}

	jobID := uuid.New().String()
	err = deployV1.Jobs().Create(jobID, auth.UserID, model.JobDeleteVM, version.V2, map[string]interface{}{
		"id": vm.ID,
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
// @Summary Updates a VM
// @Description Updates a VM
// @Tags VM
// @Accept  json
// @Produce  json
// @Param Authorization header string true "Bearer token"
// @Param vmId path string true "VM ID"
// @Param body body body.VmUpdate true "VM update"
// @Success 200 {object} body.VmUpdated
// @Failure 400 {object} sys.ErrorResponse
// @Failure 401 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 423 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /vm/{vmId} [post]
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

	var vm *model.VM
	if requestBody.TransferCode != nil {
		vm, err = deployV2.VMs().Get("", opts.GetOpts{TransferCode: requestBody.TransferCode})
		if err != nil {
			context.ServerError(err, v1.InternalError)
			return
		}

		if requestBody.OwnerID == nil {
			requestBody.OwnerID = &auth.UserID
		}

	} else {
		vm, err = deployV2.VMs().Get(requestURI.VmID, opts.GetOpts{Shared: true})
		if err != nil {
			context.ServerError(err, v1.InternalError)
			return
		}
	}

	if vm == nil {
		context.NotFound("VM not found")
		return
	}

	//if requestBody.OwnerID != nil {
	//	if *requestBody.OwnerID == "" {
	//		err = deployV2.VMs().ClearUpdateOwner(vm.ID)
	//		if err != nil {
	//			if errors.Is(err, sErrors.VmNotFoundErr) {
	//				context.NotFound("VM not found")
	//				return
	//			}
	//
	//			context.ServerError(err, v1.InternalError)
	//			return
	//		}
	//
	//		context.Ok(body.VmUpdated{
	//			ID: vm.ID,
	//		})
	//		return
	//	}
	//
	//	if *requestBody.OwnerID == vm.OwnerID {
	//		context.UserError("Owner already set")
	//		return
	//	}
	//
	//	exists, err := deployV1.Users().Exists(*requestBody.OwnerID)
	//	if err != nil {
	//		context.ServerError(err, v1.InternalError)
	//		return
	//	}
	//
	//	if !exists {
	//		context.UserError("User not found")
	//		return
	//	}
	//
	//	jobID, err := deployV2.VMs().UpdateOwnerSetup(vm.ID, &body.VmUpdateOwner{
	//		NewOwnerID:   *requestBody.OwnerID,
	//		OldOwnerID:   vm.OwnerID,
	//		TransferCode: requestBody.TransferCode,
	//	})
	//	if err != nil {
	//		if errors.Is(err, sErrors.VmNotFoundErr) {
	//			context.NotFound("VM not found")
	//			return
	//		}
	//
	//		if errors.Is(err, sErrors.InvalidTransferCodeErr) {
	//			context.Forbidden("Bad transfer code")
	//			return
	//		}
	//
	//		context.ServerError(err, v1.InternalError)
	//		return
	//	}
	//
	//	context.Ok(body.VmUpdated{
	//		ID:    vm.ID,
	//		JobID: jobID,
	//	})
	//	return
	//}

	//if requestBody.GpuID != nil {
	//	updateGPU(&context, &requestBody, deployV1, vm)
	//	return
	//}

	//if requestBody.Name != nil {
	//	available, err := deployV1.VMs().NameAvailable(*requestBody.Name)
	//	if err != nil {
	//		context.ServerError(err, v1.InternalError)
	//		return
	//	}
	//
	//	if !available {
	//		context.UserError("Name already taken")
	//		return
	//	}
	//}

	//if requestBody.Ports != nil {
	//	for _, port := range *requestBody.Ports {
	//		if port.HttpProxy != nil {
	//			available, err := deployV1.VMs().HttpProxyNameAvailable(requestURI.VmID, port.HttpProxy.Name)
	//			if err != nil {
	//				context.ServerError(err, v1.InternalError)
	//				return
	//			}
	//
	//			if !available {
	//				context.UserError("Http proxy name already taken")
	//				return
	//			}
	//		}
	//	}
	//}

	err = deployV2.VMs().CheckQuota(auth.UserID, vm.ID, &auth.GetEffectiveRole().Quotas, opts.QuotaOpts{Update: &requestBody})
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
	err = deployV1.Jobs().Create(jobID, auth.UserID, model.JobUpdateVM, version.V2, map[string]interface{}{
		"id":     vm.ID,
		"params": requestBody,
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
