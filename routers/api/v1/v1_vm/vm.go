package v1_vm

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/query"
	"go-deploy/models/dto/uri"
	jobModels "go-deploy/models/sys/job"
	vmModels "go-deploy/models/sys/vm"
	zoneModels "go-deploy/models/sys/zone"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/job_service"
	"go-deploy/service/user_service"
	"go-deploy/service/vm_service"
	"go-deploy/service/vm_service/client"
	"go-deploy/service/zone_service"
	"go-deploy/utils"
)

// List
// @Summary Get list of VMs
// @Description Get list of VMs
// @Tags VM
// @Accept  json
// @Produce  json
// @Param all query bool false "Get all"
// @Param userId query string false "Filter by user id"
// @Param page query int false "Page number"
// @Param pageSize query int false "Number of items per page"
// @Success 200 {array} body.VmRead
// @Failure 500 {object} sys.ErrorResponse
// @Router /vm [get]
func List(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.VmList
	if err := context.GinContext.Bind(&requestQuery); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	vsc := vm_service.New().WithAuth(auth)

	var userID string
	if requestQuery.UserID != nil {
		userID = *requestQuery.UserID
	} else if !requestQuery.All {
		userID = auth.UserID
	}

	vms, err := vsc.List(&client.ListOptions{
		Pagination: service.GetOrDefault(requestQuery.Pagination),
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
		connectionString, _ := vsc.GetConnectionString(vm.ID)

		var gpuRead *body.GpuRead
		if gpu := vm.GetGpu(); gpu != nil {
			gpuDTO := gpu.ToDTO(true)
			gpuRead = &gpuDTO
		}

		mapper, err := vsc.GetExternalPortMapper(vm.ID)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to get external port mapper for vm when listing. details: %w", err))
			continue
		}

		dtoVMs[i] = vm.ToDTO(vm.StatusMessage, connectionString, getTeamIDs(vm.ID, auth), gpuRead, mapper)
	}

	context.Ok(dtoVMs)
}

// Get
// @Summary Get VM by id
// @Description Get VM by id
// @Tags VM
// @Accept  json
// @Produce  json
// @Param vmId path string true "VM ID"
// @Success 200 {object} body.VmRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /vm/{vmId} [get]
func Get(c *gin.Context) {
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

	vsc := vm_service.New().WithAuth(auth)

	vm, err := vsc.Get(requestURI.VmID, client.GetOptions{Shared: true})
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if vm == nil {
		context.NotFound("VM not found")
		return
	}

	connectionString, _ := vsc.GetConnectionString(requestURI.VmID)
	var gpuRead *body.GpuRead
	if gpu := vm.GetGpu(); gpu != nil {
		gpuDTO := gpu.ToDTO(true)
		gpuRead = &gpuDTO
	}

	mapper, err := vsc.GetExternalPortMapper(vm.ID)
	if err != nil {
		utils.PrettyPrintError(fmt.Errorf("failed to get external port mapper for vm %s. details: %w", vm.ID, err))
	}

	context.Ok(vm.ToDTO(vm.StatusMessage, connectionString, getTeamIDs(vm.ID, auth), gpuRead, mapper))
}

// Create
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
// @Router /vm [post]
func Create(c *gin.Context) {
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

	vsc := vm_service.New().WithAuth(auth)

	unique, err := vm_service.NameAvailable(requestBody.Name)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if !unique {
		context.UserError("VM already exists")
		return
	}

	if requestBody.Zone != nil {
		zone := zone_service.New().WithAuth(auth).Get(*requestBody.Zone, zoneModels.TypeVM)
		if zone == nil {
			context.NotFound("Zone not found")
			return
		}
	}

	err = vsc.CheckQuota("", auth.UserID, &auth.GetEffectiveRole().Quotas, client.QuotaOptions{Create: &requestBody})
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
	err = job_service.New().Create(jobID, auth.UserID, jobModels.TypeCreateVM, map[string]interface{}{
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

// Delete
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
func Delete(c *gin.Context) {
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

	vsc := vm_service.New().WithAuth(auth)

	vm, err := vsc.Get(requestURI.VmID, client.GetOptions{Shared: true})
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

	err = vsc.StartActivity(vm.ID, vmModels.ActivityBeingDeleted)
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

		context.ServerError(err, v1.InternalError)
		return
	}

	jobID := uuid.New().String()
	err = job_service.New().Create(jobID, auth.UserID, jobModels.TypeDeleteVM, map[string]interface{}{
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

// Update
// @Summary Update VM
// @Description Update VM
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
func Update(c *gin.Context) {
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

	vsc := vm_service.New().WithAuth(auth)

	var vm *vmModels.VM
	if requestBody.TransferCode != nil {
		vm, err = vsc.Get("", client.GetOptions{TransferCode: requestBody.TransferCode})
		if err != nil {
			context.ServerError(err, v1.InternalError)
			return
		}

		if requestBody.OwnerID == nil {
			requestBody.OwnerID = &auth.UserID
		}

	} else {
		vm, err = vsc.Get(requestURI.VmID, client.GetOptions{Shared: true})
		if err != nil {
			context.ServerError(err, v1.InternalError)
			return
		}
	}

	if vm == nil {
		context.NotFound("VM not found")
		return
	}

	if requestBody.OwnerID != nil {
		if *requestBody.OwnerID == "" {
			err = vsc.ClearUpdateOwner(vm.ID)
			if err != nil {
				if errors.Is(err, sErrors.VmNotFoundErr) {
					context.NotFound("VM not found")
					return
				}

				context.ServerError(err, v1.InternalError)
				return
			}

			context.Ok(body.VmUpdated{
				ID: vm.ID,
			})
			return
		}

		if *requestBody.OwnerID == vm.OwnerID {
			context.UserError("Owner already set")
			return
		}

		exists, err := user_service.New().Exists(*requestBody.OwnerID)
		if err != nil {
			context.ServerError(err, v1.InternalError)
			return
		}

		if !exists {
			context.UserError("User not found")
			return
		}

		jobID, err := vsc.UpdateOwnerSetup(vm.ID, &body.VmUpdateOwner{
			NewOwnerID:   *requestBody.OwnerID,
			OldOwnerID:   vm.OwnerID,
			TransferCode: requestBody.TransferCode,
		})
		if err != nil {
			if errors.Is(err, sErrors.VmNotFoundErr) {
				context.NotFound("VM not found")
				return
			}

			if errors.Is(err, sErrors.InvalidTransferCodeErr) {
				context.Forbidden("Bad transfer code")
				return
			}

			context.ServerError(err, v1.InternalError)
			return
		}

		context.Ok(body.VmUpdated{
			ID:    vm.ID,
			JobID: jobID,
		})
		return
	}

	if requestBody.GpuID != nil {
		updateGPU(&context, &requestBody, auth, vm)
		return
	}

	if requestBody.Name != nil {
		available, err := vm_service.NameAvailable(*requestBody.Name)
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
				available, err := vm_service.HttpProxyNameAvailable(requestURI.VmID, port.HttpProxy.Name)
				if err != nil {
					context.ServerError(err, v1.InternalError)
					return
				}

				if !available {
					context.UserError("Http proxy name already taken")
					return
				}
			}
		}
	}

	err = vsc.CheckQuota(auth.UserID, vm.ID, &auth.GetEffectiveRole().Quotas, client.QuotaOptions{Update: &requestBody})
	if err != nil {
		var quotaExceededErr sErrors.QuotaExceededError
		if errors.As(err, &quotaExceededErr) {
			context.Forbidden(quotaExceededErr.Error())
			return
		}

		context.ServerError(err, v1.InternalError)
		return
	}

	err = vsc.StartActivity(vm.ID, vmModels.ActivityUpdating)
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

		context.ServerError(err, v1.InternalError)
		return
	}

	jobID := uuid.New().String()
	err = job_service.New().Create(jobID, auth.UserID, jobModels.TypeUpdateVM, map[string]interface{}{
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

// getTeamIDs returns a list of team IDs that the user is a member of and has access to the VM
// TODO: this currently fetches the entire team models from the database, but ideally, this should
// 1. be cached 2. only fetch the team IDs using projection
func getTeamIDs(resourceID string, auth *service.AuthInfo) []string {
	teams, err := user_service.New().ListTeams(user_service.ListTeamsOpts{ResourceID: resourceID, UserID: auth.UserID})

	if err != nil {
		return []string{}
	}

	teamIDs := make([]string, len(teams))
	for idx, team := range teams {
		teamIDs[idx] = team.ID
	}

	return teamIDs
}
