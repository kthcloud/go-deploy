package v1_user

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/query"
	"go-deploy/models/dto/uri"
	teamModels "go-deploy/models/sys/user/team"
	"go-deploy/pkg/app/status_codes"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/deployment_service"
	"go-deploy/service/user_service"
	"go-deploy/service/vm_service"
	"go-deploy/utils"
	"net/http"
)

// GetTeam godoc
// @Summary Get team
// @Description Get team
// @Tags Team
// @Accept json
// @Produce json
// @Param teamId path string true "Team ID"
// @Success 200 {object} body.TeamRead
// @Failure 400 {object} sys.ErrorResponse
func GetTeam(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.TeamUpdate
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	team, err := user_service.GetTeamByIdAuth(requestURI.TeamID, auth)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get team: %s", err))
		return
	}

	context.JSONResponse(http.StatusOK, team.ToDTO(getMember, getResourceName))
}

func GetTeamList(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.TeamList
	if err := context.GinContext.ShouldBind(&requestQuery); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	teamList, err := user_service.GetTeamListAuth(requestQuery.All, requestQuery.UserID, auth, &requestQuery.Pagination)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get team list: %s", err))
		return
	}

	teamListDTO := make([]body.TeamRead, len(teamList))
	for i, team := range teamList {
		teamListDTO[i] = team.ToDTO(getMember, getResourceName)
	}

	context.JSONResponse(http.StatusOK, teamListDTO)
}

func CreateTeam(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery body.TeamCreate
	if err := context.GinContext.ShouldBindJSON(&requestQuery); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	team, err := user_service.CreateTeam(uuid.NewString(), auth.UserID, &requestQuery, auth)
	if err != nil {
		if errors.Is(err, user_service.TeamNameTakenErr) {
			context.ErrorResponse(http.StatusBadRequest, status_codes.ResourceNotAvailable, "Team name is taken")
			return
		}

		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to create team: %s", err))
		return
	}

	context.JSONResponse(http.StatusCreated, team.ToDTO(getMember, getResourceName))
}

func UpdateTeam(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.TeamUpdate
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	var requestQuery body.TeamUpdate
	if err := context.GinContext.ShouldBindJSON(&requestQuery); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	team, err := user_service.UpdateTeamAuth(requestURI.TeamID, &requestQuery, auth)
	if err != nil {
		if errors.Is(err, user_service.TeamNameTakenErr) {
			context.ErrorResponse(http.StatusBadRequest, status_codes.ResourceNotAvailable, "Team name is taken")
			return
		}

		if errors.Is(err, user_service.TeamNotFoundErr) {
			context.ErrorResponse(http.StatusNotFound, status_codes.ResourceNotFound, "Team not found")
			return
		}

		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to update team: %s", err))
		return
	}

	context.JSONResponse(http.StatusOK, team.ToDTO(getMember, getResourceName))
}

func DeleteTeam(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.TeamUpdate
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	err = user_service.DeleteTeamAuth(requestURI.TeamID, auth)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to delete team: %s", err))
		return
	}

	context.OkDeleted()
}

func getMember(member *teamModels.Member) *body.TeamMember {
	user, err := user_service.GetByID(member.ID)
	if err != nil {
		utils.PrettyPrintError(fmt.Errorf("failed to get user when getting team member for team: %s", err))
		return nil
	}

	if user == nil {
		return nil
	}

	return &body.TeamMember{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
		TeamRole: member.TeamRole,
		JoinedAt: member.JoinedAt,
	}
}

func getResourceName(resource *teamModels.Resource) *string {
	if resource == nil {
		return nil
	}

	switch resource.Type {
	case teamModels.ResourceTypeDeployment:
		d, err := deployment_service.GetByID(resource.ID)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to get deployment when getting team resource name: %s", err))
			return nil
		}

		if d == nil {
			return nil
		}

		return &d.Name
	case teamModels.ResourceTypeVM:
		vm, err := vm_service.GetByID(resource.ID)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to get vm when getting team resource name: %s", err))
			return nil
		}

		if vm == nil {
			return nil
		}

		return &vm.Name
	}

	return nil

}
