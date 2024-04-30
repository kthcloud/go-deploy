package v1

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-deploy/dto/v1/body"
	"go-deploy/dto/v1/query"
	"go-deploy/dto/v1/uri"
	"go-deploy/models/model"
	"go-deploy/models/version"
	"go-deploy/pkg/sys"
	"go-deploy/service"
	sErrors "go-deploy/service/errors"
	teamOpts "go-deploy/service/v1/teams/opts"
	v1Utils "go-deploy/service/v1/utils"
	"go-deploy/service/v1/vms/opts"
	"go-deploy/utils"
)

// GetVM
// @Summary Get VM
// @Description Get VM
// @Tags VM
// @Accept  json
// @Produce  json
// @Param vmId path string true "VM ID"
// @Success 200 {object} body.VmRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v1/vms/{vmId} [get]
func GetVM(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.VmGet
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, AuthInfoNotAvailableErr)
		return
	}

	deployV1 := service.V1(auth)

	vm, err := deployV1.VMs().Get(requestURI.VmID, opts.GetOpts{Shared: true})
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	if vm == nil {
		context.NotFound("VM not found")
		return
	}

	connectionString, _ := deployV1.VMs().GetConnectionString(requestURI.VmID)
	var gpuRead *body.GpuRead
	if gpu, _ := deployV1.VMs().GetGpuByVM(vm.ID); gpu != nil {
		gpuDTO := gpu.ToDTO(true)
		gpuRead = &gpuDTO
	}

	mapper, err := deployV1.VMs().GetExternalPortMapper(vm.ID)
	if err != nil {
		utils.PrettyPrintError(fmt.Errorf("failed to get external port mapper for vm %s. details: %w", vm.ID, err))
	}

	teamIDs, _ := deployV1.Teams().ListIDs(teamOpts.ListOpts{ResourceID: vm.ID})
	context.Ok(vm.ToDTOv1(vm.StatusMessage, connectionString, teamIDs, gpuRead, mapper))
}

// ListVMs
// @Summary List VMs
// @Description List VMs
// @Tags VM
// @Accept  json
// @Produce  json
// @Param all query bool false "GetVM all"
// @Param userId query string false "Filter by user id"
// @Param page query int false "Page number"
// @Param pageSize query int false "Number of items per page"
// @Success 200 {array} body.VmRead
// @Failure 500 {object} sys.ErrorResponse
// @Router /v1/vms [get]
func ListVMs(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.VmList
	if err := context.GinContext.ShouldBind(&requestQuery); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, AuthInfoNotAvailableErr)
		return
	}

	deployV1 := service.V1(auth)

	var userID string
	if requestQuery.UserID != nil {
		userID = *requestQuery.UserID
	} else if !requestQuery.All {
		userID = auth.UserID
	}

	vms, err := deployV1.VMs().List(opts.ListOpts{
		Pagination: v1Utils.GetOrDefaultPagination(requestQuery.Pagination),
		UserID:     &userID,
		Shared:     true,
	})
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	if vms == nil {
		context.Ok([]interface{}{})
		return
	}

	dtoVMs := make([]body.VmRead, len(vms))
	for i, vm := range vms {
		connectionString, _ := deployV1.VMs().GetConnectionString(vm.ID)

		var gpuRead *body.GpuRead
		if gpu, _ := deployV1.VMs().GetGpuByVM(vm.ID); gpu != nil {
			gpuDTO := gpu.ToDTO(true)
			gpuRead = &gpuDTO
		}

		mapper, err := deployV1.VMs().GetExternalPortMapper(vm.ID)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to get external port mapper for vm when listing. details: %w", err))
			continue
		}

		teamIDs, _ := deployV1.Teams().ListIDs(teamOpts.ListOpts{ResourceID: vm.ID})
		dtoVMs[i] = vm.ToDTOv1(vm.StatusMessage, connectionString, teamIDs, gpuRead, mapper)
	}

	context.Ok(dtoVMs)
}

// CreateVM
// @Summary Create VM
// @Description Create VM
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
// @Router /v1/vms [post]
func CreateVM(c *gin.Context) {
	context := sys.NewContext(c)

	var requestBody body.VmCreate
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

	unique, err := deployV1.VMs().NameAvailable(requestBody.Name)
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	if !unique {
		context.UserError("VM already exists")
		return
	}

	if requestBody.Zone != nil {
		zone := deployV1.Zones().GetLegacy(*requestBody.Zone)
		if zone == nil {
			context.NotFound("Zone not found")
			return
		}
	}

	err = deployV1.VMs().CheckQuota("", auth.UserID, &auth.GetEffectiveRole().Quotas, opts.QuotaOpts{Create: &requestBody})
	if err != nil {
		var quotaExceedErr sErrors.QuotaExceededError
		if errors.As(err, &quotaExceedErr) {
			context.Forbidden(quotaExceedErr.Error())
			return
		}

		context.ServerError(err, InternalError)
		return
	}

	vmID := uuid.New().String()
	jobID := uuid.New().String()
	err = deployV1.Jobs().Create(jobID, auth.UserID, model.JobCreateVM, version.V1, map[string]interface{}{
		"id":       vmID,
		"ownerId":  auth.UserID,
		"params":   requestBody,
		"authInfo": auth,
	})
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	context.Ok(body.VmCreated{
		ID:    vmID,
		JobID: jobID,
	})
}

// UpdateVM
// @Summary UpdateVM VM
// @Description UpdateVM VM
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
// @Router /v1/vms/{vmId} [post]
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
		context.ServerError(err, AuthInfoNotAvailableErr)
		return
	}

	deployV1 := service.V1(auth)

	vm, err := deployV1.VMs().Get(requestURI.VmID, opts.GetOpts{Shared: true})
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	if vm == nil {
		context.NotFound("VM not found")
		return
	}

	if requestBody.GpuID != nil {
		updateGPU(&context, &requestBody, deployV1, vm)
		return
	}

	if requestBody.Name != nil {
		available, err := deployV1.VMs().NameAvailable(*requestBody.Name)
		if err != nil {
			context.ServerError(err, InternalError)
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
				available, err := deployV1.VMs().HttpProxyNameAvailable(requestURI.VmID, port.HttpProxy.Name)
				if err != nil {
					context.ServerError(err, InternalError)
					return
				}

				if !available {
					context.UserError("Http proxy name already taken")
					return
				}
			}
		}
	}

	err = deployV1.VMs().CheckQuota(vm.ID, auth.UserID, &auth.GetEffectiveRole().Quotas, opts.QuotaOpts{Update: &requestBody})
	if err != nil {
		var quotaExceededErr sErrors.QuotaExceededError
		if errors.As(err, &quotaExceededErr) {
			context.Forbidden(quotaExceededErr.Error())
			return
		}

		context.ServerError(err, InternalError)
		return
	}

	err = deployV1.VMs().StartActivity(vm.ID, model.ActivityUpdating)
	if err != nil {
		var failedToStartActivityErr sErrors.FailedToStartActivityError
		if errors.As(err, &failedToStartActivityErr) {
			context.Locked(failedToStartActivityErr.Error())
			return
		}

		if errors.Is(err, sErrors.VmNotFoundErr) {
			context.NotFound("Deployment not found")
			return
		}

		context.ServerError(err, InternalError)
		return
	}

	jobID := uuid.New().String()
	err = deployV1.Jobs().Create(jobID, auth.UserID, model.JobUpdateVM, version.V1, map[string]interface{}{
		"id":       vm.ID,
		"params":   requestBody,
		"authInfo": auth,
	})

	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	context.Ok(body.VmUpdated{
		ID:    vm.ID,
		JobID: &jobID,
	})
}

// DeleteVM
// @Summary DeleteVM VM
// @Description DeleteVM VM
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
// @Router /v1/vms/{vmId} [delete]
func DeleteVM(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.VmDelete
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, AuthInfoNotAvailableErr)
		return
	}

	deployV1 := service.V1(auth)

	vm, err := deployV1.VMs().Get(requestURI.VmID, opts.GetOpts{Shared: true})
	if err != nil {
		context.ServerError(err, InternalError)
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

	err = deployV1.VMs().StartActivity(vm.ID, model.ActivityBeingDeleted)
	if err != nil {
		var failedToStartActivityErr sErrors.FailedToStartActivityError
		if errors.As(err, &failedToStartActivityErr) {
			context.Locked(failedToStartActivityErr.Error())
			return
		}

		if errors.Is(err, sErrors.VmNotFoundErr) {
			context.NotFound("Deployment not found")
			return
		}

		context.ServerError(err, InternalError)
		return
	}

	jobID := uuid.New().String()
	err = deployV1.Jobs().Create(jobID, auth.UserID, model.JobDeleteVM, version.V1, map[string]interface{}{
		"id":       vm.ID,
		"authInfo": auth,
	})
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	context.Ok(body.VmDeleted{
		ID:    vm.ID,
		JobID: jobID,
	})
}
