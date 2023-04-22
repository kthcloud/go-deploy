package v1_vm

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-deploy/models/dto"
	"go-deploy/pkg/app"
	"go-deploy/pkg/status_codes"
	"go-deploy/pkg/validator"
	"go-deploy/service/user_info_service"
	"go-deploy/service/vm_service"
	"net/http"
	"strconv"
)

func getAllVMs(context *app.ClientContext) {
	vms, _ := vm_service.GetAll()

	dtoVMs := make([]dto.VmRead, len(vms))
	for i, vm := range vms {
		connectionString, _ := vm_service.GetConnectionString(&vm)
		dtoVMs[i] = vm.ToDto(vm.StatusMessage, connectionString)
	}

	context.JSONResponse(http.StatusOK, dtoVMs)
}

func GetMany(c *gin.Context) {
	context := app.NewContext(c)

	rules := validator.MapData{
		"all": []string{"bool"},
	}

	validationErrors := context.ValidateQueryParams(&rules)
	if len(validationErrors) > 0 {
		context.ResponseValidationError(validationErrors)
		return
	}

	token, err := context.GetKeycloakToken()
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}
	userID := token.Sub

	// might want to check if userID is allowed to get all...
	wantAll, _ := strconv.ParseBool(context.GinContext.Query("all"))
	if wantAll {
		getAllVMs(&context)
		return
	}

	vms, _ := vm_service.GetByOwnerID(userID)
	if vms == nil {
		context.JSONResponse(200, []interface{}{})
		return
	}

	dtoVMs := make([]dto.VmRead, len(vms))
	for i, vm := range vms {
		connectionString, _ := vm_service.GetConnectionString(&vm)
		dtoVMs[i] = vm.ToDto(vm.StatusMessage, connectionString)
	}

	context.JSONResponse(200, dtoVMs)
}

func Get(c *gin.Context) {
	context := app.NewContext(c)

	rules := validator.MapData{
		"vmId": []string{"required", "uuid_v4"},
	}

	validationErrors := context.ValidateParams(&rules)

	if len(validationErrors) > 0 {
		context.ResponseValidationError(validationErrors)
		return
	}

	token, err := context.GetKeycloakToken()
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
	}
	vmID := context.GinContext.Param("vmId")
	userID := token.Sub

	vm, _ := vm_service.GetByID(userID, vmID)

	if vm == nil {
		context.NotFound()
		return
	}

	connectionString, _ := vm_service.GetConnectionString(vm)
	context.JSONResponse(200, vm.ToDto(vm.StatusMessage, connectionString))
}

func Create(c *gin.Context) {
	context := app.NewContext(c)

	bodyRules := validator.MapData{
		"name": []string{
			"required",
			"regex:^[a-zA-Z]([a-zA-Z0-9-]*[a-zA-Z0-9])?([a-zA-Z]([a-zA-Z0-9-]*[a-zA-Z0-9])?)*$",
			"min:3",
			"max:30",
		},
	}

	messages := validator.MapData{
		"name": []string{
			"required:Name is required",
			"regexp:Name must follow RFC 1035 and must not include any dots",
			"min:Name must be between 3-30 characters",
			"max:Name must be between 3-30 characters",
		},
	}

	var requestBody dto.VmCreate
	validationErrors := context.ValidateJSONCustomMessages(&bodyRules, &messages, &requestBody)
	if len(validationErrors) > 0 {
		context.ResponseValidationError(validationErrors)
		return
	}

	token, err := context.GetKeycloakToken()
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	userInfo, err := user_info_service.GetByToken(token)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	if userInfo.VmQuota == 0 {
		context.ErrorResponse(http.StatusUnauthorized, status_codes.Error, "User is not allowed to create vms")
		return
	}

	userID := token.Sub

	exists, vm, err := vm_service.Exists(requestBody.Name)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.ResourceValidationFailed, "Failed to validate")
		return
	}

	// temporary, this should be handled by the validator
	//sshPublicKey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDHJ+XBsrl/eUIcDHJf8tA22wgocd8+r6zH47VzNl1M9Ri6tlgEeH13O6b6yO7W38wjx+Ftcv+9P93XeAf3N8h78JuvWlb8Q/xPMFZxSePRpiYtDqCR3ClEZ8KkKYgS/APeybZy9fNH8JduuvSAp5FkDVnW8VZfUpKUm0w3Ka32jtAwAOb5ghIdSc35hL37hLnB0PVz9q3f5OD2g1bEx187IunDrQYkp8YVDPxLI0qc7iARFYpEvNfTiRaWMRywAd7ANa4LQYc4KyWZxEsAZ+pjdOsp7WkaHrbeBypLFh+9+3nEYcT4CTj9r0jIM2e8m1Y7t79heMy/AqQF2FsaOvFuow70RjFmrIrC2Z/AylJDkYtcgy8cxafviISwlplgQ0XQsTsc4OAZAGyLvNHgUh3VeArXa4YczlDSK+IlFUwDr87r7MPoLETO9RhraA98ksHUzPQ0/J1NbjwB+vMCAAak1Zv4MIJLdX2XPjzDrUvPdnkyt4OLZD++RVM3EuOqDDM= cloud@se-flem-001"
	sshPublicKey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDBjv2iP3pR5vZO/ESY6ZZdsk4ZxdtZiQ/OiPbztuShFdj6l7+Azlga9/n7TLTMVl1HjJT1OfJKNV7Dhfh7gzf0LgjM4yPh/Ovcwg3/K1ycGOBFKAo8KhhOq2kpuKOvy0Ug2wFPngZxQfFUM0a1t4MuSjiX4CP5eUPTc8gOcetf8konKjbRk5h/gYzBqH6edpBz4RvgskuCCKJSQ3h7cfEflhvItUQ4yHFihJewPQLLfhZDuMoc9zHJObqOzAhN7NJ0kyNrNo45QtxarTuavKZZZ1hGrNL9tOrebrf1OU2jtMmag12MUcKNa4sBOGe6J3qII8tsIcEumSqtWYywyK9Wa8kNkKItHMLognXKc99bDZRu3yBSAiOsfC8197c0mqGfd4PbILXmOfkEV0aRIL0ElUk6EmKl1+KvQ0lYDb8SIopl5UAyDObwDFfdAKDORGddUqn2onDi+vj+hQwwcbNFTE+H6CCra2JLmIa/XAGpZ26SQTK1/Shp6ALw4ZJv1jc= ownem@Home"

	if exists {
		if vm.OwnerID != userID {
			context.ErrorResponse(http.StatusBadRequest, status_codes.ResourceAlreadyExists, "Resource already exists")
			return
		}
		if vm.BeingDeleted {
			context.ErrorResponse(http.StatusLocked, status_codes.ResourceBeingDeleted, "Resource is currently being deleted")
			return
		}
		vm_service.Create(vm.ID, requestBody.Name, sshPublicKey, userID)
		context.JSONResponse(http.StatusCreated, dto.VmCreated{ID: vm.ID})
		return
	}

	vmCount, err := vm_service.GetCount(userID)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	if vmCount >= userInfo.VmQuota {
		context.ErrorResponse(http.StatusUnauthorized, status_codes.Error, fmt.Sprintf("User is not allowed to create more than %d vms", userInfo.VmQuota))
		return
	}

	vmID := uuid.New().String()
	vm_service.Create(vmID, requestBody.Name, sshPublicKey, userID)
	context.JSONResponse(http.StatusCreated, dto.VmCreated{ID: vmID})
}

func Delete(c *gin.Context) {
	context := app.NewContext(c)

	rules := validator.MapData{
		"vmId": []string{
			"required",
			"uuid_v4",
		},
	}

	validationErrors := context.ValidateParams(&rules)
	if len(validationErrors) > 0 {
		context.ResponseValidationError(validationErrors)
		return
	}

	token, err := context.GetKeycloakToken()
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}
	userID := token.Sub
	vmID := context.GinContext.Param("vmId")

	current, err := vm_service.GetByID(userID, vmID)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.ResourceValidationFailed, "Failed to validate")
		return
	}

	if current == nil {
		context.NotFound()
		return
	}

	if current.BeingCreated {
		context.ErrorResponse(http.StatusLocked, status_codes.ResourceBeingCreated, "Resource is currently being created")
		return
	}

	if !current.BeingDeleted {
		_ = vm_service.MarkBeingDeleted(current.ID)
	}

	vm_service.Delete(current.Name)

	context.OkDeleted()
}
