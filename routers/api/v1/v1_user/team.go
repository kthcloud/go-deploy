package v1_user

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/query"
	"go-deploy/models/dto/uri"
	teamModels "go-deploy/models/sys/team"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/deployment_service"
	dClient "go-deploy/service/deployment_service/client"
	"go-deploy/service/user_service"
	"go-deploy/service/vm_service"
	vClient "go-deploy/service/vm_service/client"
	"go-deploy/utils"
	"net/http"
	"time"
)

// ListTeams godoc
// @Summary Get team list
// @Description Get team list
// @Tags Team
// @Accept json
// @Produce json
// @Param all query bool false "All teams"
// @Param userId query string false "User ID"
// @Param page query int false "Page"
// @Param pageSize query int false "Page Size"
// @Success 200 {array} body.TeamRead
// @Failure 400 {object} body.BindingError
// @Failure 500 {object} sys.ErrorResponse
// @Router /teams [get]
func ListTeams(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.TeamList
	if err := context.GinContext.ShouldBind(&requestQuery); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	teamList, err := user_service.ListTeamsAuth(requestQuery.All, requestQuery.UserID, auth, &requestQuery.Pagination)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	teamListDTO := make([]body.TeamRead, len(teamList))
	for i, team := range teamList {
		teamListDTO[i] = team.ToDTO(getMember, getResourceName)
	}

	context.Ok(teamListDTO)
}

// GetTeam godoc
// @Summary Get team
// @Description Get team
// @Tags Team
// @Accept json
// @Produce json
// @Param teamId path string true "Team ID"
// @Success 200 {object} body.TeamRead
// @Failure 400 {object} body.BindingError
// @Failure 500 {object} sys.ErrorResponse
// @Router /teams/{teamId} [get]
func GetTeam(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.TeamGet
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	team, err := user_service.GetTeamByIdAuth(requestURI.TeamID, auth)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	context.Ok(team.ToDTO(getMember, getResourceName))
}

// CreateTeam godoc
// @Summary Create team
// @Description Create team
// @Tags Team
// @Accept json
// @Produce json
// @Param body body body.TeamCreate true "Team"
// @Success 200 {object} body.TeamRead
// @Failure 400 {object} body.BindingError
// @Failure 500 {object} sys.ErrorResponse
// @Router /teams [post]
func CreateTeam(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery body.TeamCreate
	if err := context.GinContext.ShouldBindJSON(&requestQuery); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	team, err := user_service.CreateTeam(uuid.NewString(), auth.UserID, &requestQuery, auth)
	if err != nil {
		if errors.Is(err, user_service.TeamNameTakenErr) {
			context.UserError("Team name is taken")
			return
		}

		context.ServerError(err, v1.InternalError)
		return
	}

	context.JSONResponse(http.StatusCreated, team.ToDTO(getMember, getResourceName))
}

// UpdateTeam godoc
// @Summary Update team
// @Description Update team
// @Tags Team
// @Accept json
// @Produce json
// @Param teamId path string true "Team ID"
// @Param body body body.TeamUpdate true "Team"
// @Success 200 {object} body.TeamRead
// @Failure 400 {object} body.BindingError
// @Failure 500 {object} sys.ErrorResponse
// @Router /teams/{teamId} [post]
func UpdateTeam(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.TeamUpdate
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	var requestQueryJoin body.TeamJoin
	if err := context.GinContext.ShouldBindBodyWith(&requestQueryJoin, binding.JSON); err == nil {
		joinTeam(context, requestURI.TeamID, &requestQueryJoin)
		return
	}

	var requestQuery body.TeamUpdate
	if err := context.GinContext.ShouldBindBodyWith(&requestQuery, binding.JSON); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	updated, err := user_service.UpdateTeamAuth(requestURI.TeamID, &requestQuery, auth)
	if err != nil {
		if errors.Is(err, user_service.TeamNameTakenErr) {
			context.UserError("Team name is taken")
			return
		}

		context.ServerError(err, v1.InternalError)
		return
	}

	if updated == nil {
		context.NotFound("Team not found")
		return
	}

	context.JSONResponse(http.StatusOK, updated.ToDTO(getMember, getResourceName))
}

// DeleteTeam godoc
// @Summary Delete team
// @Description Delete team
// @Tags Team
// @Accept json
// @Produce json
// @Param teamId path string true "Team ID"
// @Success 204 "No Content"
// @Failure 400 {object} body.BindingError
// @Failure 500 {object} sys.ErrorResponse
// @Router /teams/{teamId} [delete]
func DeleteTeam(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.TeamUpdate
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	err = user_service.DeleteTeamAuth(requestURI.TeamID, auth)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	context.OkNoContent()
}

func getMember(member *teamModels.Member) *body.TeamMember {
	user, err := user_service.Get(member.ID)
	if err != nil {
		utils.PrettyPrintError(fmt.Errorf("failed to get user when getting team member for team: %s", err))
		return nil
	}

	if user == nil {
		return nil
	}

	var joinedAt *time.Time
	if !member.JoinedAt.IsZero() {
		joinedAt = &member.JoinedAt
	}

	var addedAt *time.Time
	if !member.AddedAt.IsZero() {
		addedAt = &member.AddedAt
	}

	return &body.TeamMember{
		ID:           user.ID,
		Username:     user.Username,
		Email:        user.Email,
		TeamRole:     member.TeamRole,
		JoinedAt:     joinedAt,
		AddedAt:      addedAt,
		MemberStatus: member.MemberStatus,
	}
}

func getResourceName(resource *teamModels.Resource) *string {
	if resource == nil {
		return nil
	}

	switch resource.Type {
	case teamModels.ResourceTypeDeployment:
		d, err := deployment_service.New().Get(resource.ID, &dClient.GetOptions{Shared: true})
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to get deployment when getting team resource name: %s", err))
			return nil
		}

		if d == nil {
			return nil
		}

		return &d.Name
	case teamModels.ResourceTypeVM:
		vm, err := vm_service.New().Get(resource.ID, &vClient.GetOptions{Shared: true})
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

func joinTeam(context sys.ClientContext, id string, requestBody *body.TeamJoin) {
	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
	}

	team, err := user_service.JoinTeam(id, requestBody, auth)
	if err != nil {
		if errors.Is(err, user_service.NotInvitedErr) {
			context.UserError("User not invited to team")
			return
		}

		if errors.Is(err, user_service.BadInviteCodeErr) {
			context.Forbidden("Bad invite code")
			return
		}

		context.ServerError(err, v1.InternalError)
		return
	}

	if team == nil {
		context.NotFound("Team not found")
		return
	}

	context.JSONResponse(http.StatusCreated, team.ToDTO(getMember, getResourceName))
}
