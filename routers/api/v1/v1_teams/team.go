package v1_teams

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"go-deploy/models/dto/v1/body"
	"go-deploy/models/dto/v1/query"
	"go-deploy/models/dto/v1/uri"
	teamModels "go-deploy/models/sys/team"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service"
	sErrors "go-deploy/service/errors"
	v12 "go-deploy/service/v1/common"
	dClient "go-deploy/service/v1/deployments/opts"
	"go-deploy/service/v1/teams/opts"
	vClient "go-deploy/service/v1/vms/opts"
	"go-deploy/utils"
	"net/http"
	"time"
)

// List godoc
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
func List(c *gin.Context) {
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

	var userID string
	if requestQuery.UserID != nil {
		userID = *requestQuery.UserID
	} else if !requestQuery.All {
		userID = auth.UserID
	}

	teamList, err := service.V1(auth).Teams().List(opts.ListOpts{
		Pagination: v12.GetOrDefaultPagination(requestQuery.Pagination),
		UserID:     userID,
	})
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

// Get godoc
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
func Get(c *gin.Context) {
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

	team, err := service.V1(auth).Teams().Get(requestURI.TeamID)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	context.Ok(team.ToDTO(getMember, getResourceName))
}

// Create godoc
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
func Create(c *gin.Context) {
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

	team, err := service.V1(auth).Teams().Create(uuid.NewString(), auth.UserID, &requestQuery)
	if err != nil {
		if errors.Is(err, sErrors.TeamNameTakenErr) {
			context.UserError("Team name is taken")
			return
		}

		context.ServerError(err, v1.InternalError)
		return
	}

	context.JSONResponse(http.StatusCreated, team.ToDTO(getMember, getResourceName))
}

// Update godoc
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
func Update(c *gin.Context) {
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

	updated, err := service.V1(auth).Teams().Update(requestURI.TeamID, &requestQuery)
	if err != nil {
		if errors.Is(err, sErrors.TeamNameTakenErr) {
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

// Delete godoc
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
func Delete(c *gin.Context) {
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

	err = service.V1(auth).Teams().Delete(requestURI.TeamID)
	if err != nil {
		if errors.Is(err, sErrors.TeamNotFoundErr) {
			context.NotFound("Team not found")
			return
		}

		context.ServerError(err, v1.InternalError)
		return
	}

	context.OkNoContent()
}

// joinTeam is an alternate entrypoint for Update that allows a user to join a team
// It is called if a body.TeamJoin is passed in the request body, instead of a body.TeamUpdate
func joinTeam(context sys.ClientContext, id string, requestBody *body.TeamJoin) {
	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
	}

	team, err := service.V1(auth).Teams().Join(id, requestBody)
	if err != nil {
		if errors.Is(err, sErrors.NotInvitedErr) {
			context.UserError("User not invited to team")
			return
		}

		if errors.Is(err, sErrors.BadInviteCodeErr) {
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

// getMember is a helper function for converting a team member to a team member DTO
func getMember(member *teamModels.Member) *body.TeamMember {
	user, err := service.V1().Users().Get(member.ID)
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

// getResourceName is a helper function for converting a team resource to a resource name
// It checks the resource type and gets the resource name from the appropriate service
func getResourceName(resource *teamModels.Resource) *string {
	if resource == nil {
		return nil
	}

	deployV1 := service.V1()

	switch resource.Type {
	case teamModels.ResourceTypeDeployment:
		d, err := deployV1.Deployments().Get(resource.ID, dClient.GetOpts{Shared: true})
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("failed to get deployment when getting team resource name: %s", err))
			return nil
		}

		if d == nil {
			return nil
		}

		return &d.Name
	case teamModels.ResourceTypeVM:
		vm, err := deployV1.VMs().Get(resource.ID, vClient.GetOpts{Shared: true})
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
