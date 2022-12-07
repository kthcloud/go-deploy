package v1

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-deploy/models/dto"
	"go-deploy/pkg/app"
	"go-deploy/pkg/status_codes"
	"go-deploy/pkg/validator"
	"go-deploy/service/project_service"
	"net/http"
)

func GetAllProjects(c *gin.Context) {
	context := app.NewContext(c)
	projects, _ := project_service.GetAll()

	dtoProjects := make([]dto.Project, len(projects))
	for i, project := range projects {
		dtoProjects[i] = project.ToDto()
	}

	context.JSONResponse(http.StatusOK, dtoProjects)
}

func GetProjectsByOwnerID(c *gin.Context) {
	context := app.NewContext(c)

	rules := validator.MapData{
		"userId": []string{"required", "uuid_v4"},
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

	userId := context.GinContext.Param("userId")

	if userId != token.Sub {
		context.Unauthorized()
		return
	}

	projects, _ := project_service.GetByOwner(userId)

	if projects == nil {
		context.JSONResponse(200, []interface{}{})
		return
	}

	dtoProjects := make([]dto.Project, len(projects))
	for i, project := range projects {
		dtoProjects[i] = project.ToDto()
	}

	context.JSONResponse(200, dtoProjects)
}

func GetProject(c *gin.Context) {
	context := app.NewContext(c)

	rules := validator.MapData{
		"projectId": []string{"required", "uuid_v4"},
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
	projectId := context.GinContext.Param("projectId")
	userId := token.Sub

	project, _ := project_service.Get(userId, projectId)

	if project == nil {
		context.NotFound()
		return
	}

	context.JSONResponse(200, project.ToDto())
}

func CreateProject(c *gin.Context) {
	context := app.NewContext(c)

	bodyRules := validator.MapData{
		"name": []string{
			"required",
			"regex:^[a-zA-Z]+$",
			"min:3",
			"max:10",
		},
	}

	messages := validator.MapData{
		"name": []string{
			"required:Project name is required",
			"regexp:Project name must be all lowercase",
			"min:Project name must be between 3-10 characters",
			"max:Project name must be between 3-10 characters",
		},
	}

	var requestBody dto.ProjectReq
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
	userId := token.Sub

	exists, project, err := project_service.Exists(requestBody.Name)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.ProjectValidationFailed, "Failed to validate project")
		return
	}

	if exists {
		if project.Owner != userId {
			context.ErrorResponse(http.StatusBadRequest, status_codes.ProjectAlreadyExists, "Project already exists")
			return
		}
		if project.BeingDeleted {
			context.ErrorResponse(http.StatusLocked, status_codes.ProjectBeingDeleted, "Project is currently being deleted")
			return
		}
		project_service.Create(project.ID, requestBody.Name, userId)
		context.JSONResponse(http.StatusCreated, dto.ProjectCreated{ID: project.ID})
		return
	}

	projectID := uuid.New().String()
	project_service.Create(projectID, requestBody.Name, userId)
	context.JSONResponse(http.StatusCreated, dto.ProjectCreated{ID: projectID})
}

func DeleteProject(c *gin.Context) {
	context := app.NewContext(c)

	rules := validator.MapData{
		"projectId": []string{
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
	userId := token.Sub
	projectId := context.GinContext.Param("projectId")

	currentProject, err := project_service.Get(userId, projectId)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.ProjectValidationFailed, "Failed to validate currentProject")
		return
	}

	if currentProject == nil {
		context.NotFound()
		return
	}

	if currentProject.BeingCreated {
		context.ErrorResponse(http.StatusLocked, status_codes.ProjectBeingCreated, "Project is currently being created")
		return
	}

	if !currentProject.BeingDeleted {
		_ = project_service.MarkBeingDeleted(currentProject.ID)
	}

	project_service.Delete(currentProject.Name)

	context.OkDeleted()
}
