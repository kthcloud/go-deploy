package v1_user

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/query"
	"go-deploy/models/dto/uri"
	"go-deploy/pkg/status_codes"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/user_service"
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

	context.JSONResponse(http.StatusOK, team.ToDTO(nil))
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

	context.JSONResponse(http.StatusOK, teamList)
}

func CreateTeam(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery body.TeamCreate
	if err := context.GinContext.ShouldBindJSON(&requestQuery); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	team, err := user_service.CreateTeam(uuid.NewString(), &requestQuery)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to create team: %s", err))
		return
	}

	context.JSONResponse(http.StatusCreated, team)
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
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to update team: %s", err))
		return
	}

	context.JSONResponse(http.StatusOK, team)
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

	context.JSONResponse(http.StatusOK, nil)
}
