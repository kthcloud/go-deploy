package v2

import (
	"errors"
	"github.com/gin-gonic/gin"
	"go-deploy/dto/v2/body"
	"go-deploy/dto/v2/query"
	"go-deploy/dto/v2/uri"
	"go-deploy/models/model"
	"go-deploy/pkg/config"
	"go-deploy/pkg/sys"
	"go-deploy/service"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/v2/users/opts"
	sUtils "go-deploy/service/v2/utils"
)

// GetUser
// @Summary Get user
// @Description Get user
// @Tags User
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param userId path string true "User ID"
// @Param discover query bool false "Discovery mode"
// @Success 200 {object}  body.UserRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/users/{userId} [get]
func GetUser(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.UserGet
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	var requestQuery query.UserGet
	if err := context.GinContext.ShouldBind(&requestQuery); err != nil {
		context.BindingError(CreateBindingError(err))
		return

	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, AuthInfoNotAvailableErr)
		return
	}

	if requestURI.UserID == "" {
		requestURI.UserID = auth.User.ID
	}

	deployV2 := service.V2(auth)

	if requestQuery.Discover {
		discover, err := deployV2.Users().Discover(opts.DiscoverOpts{
			UserID: &requestURI.UserID,
		})
		if err != nil {
			context.ServerError(err, InternalError)
			return
		}

		if len(discover) == 0 {
			context.NotFound("User not found")
			return
		}

		context.JSONResponse(200, discover[0])
		return
	}

	user, err := deployV2.Users().Get(requestURI.UserID)
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	if user == nil {
		context.NotFound("User not found")
		return
	}

	effectiveRole := config.Config.GetRole(user.EffectiveRole.Name)
	if effectiveRole == nil {
		effectiveRole = &model.Role{}
	}

	usage, _ := deployV2.Users().GetUsage(user.ID)
	context.JSONResponse(200, user.ToDTO(effectiveRole, usage, deployV2.SMs().GetUrlByUserID(user.ID)))
}

// ListUsers
// @Summary List users
// @Description List users
// @Tags User
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param all query bool false "List all"
// @Param discover query bool false "Discovery mode"
// @Param search query string false "Search query"
// @Param page query int false "Page number"
// @Param pageSize query int false "Number of items per page"
// @Success 200 {array}  body.UserRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/users [get]
func ListUsers(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.UserList
	if err := context.GinContext.ShouldBind(&requestQuery); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, AuthInfoNotAvailableErr)
		return
	}

	deployV2 := service.V2(auth)

	if requestQuery.Discover {
		userList, err := deployV2.Users().Discover(opts.DiscoverOpts{
			Search:     requestQuery.Search,
			Pagination: sUtils.GetOrDefaultPagination(requestQuery.Pagination),
		})
		if err != nil {
			context.ServerError(err, InternalError)
			return
		}

		if userList == nil {
			context.Ok([]interface{}{})
			return
		}

		context.Ok(userList)
		return
	}

	userList, err := deployV2.Users().List(opts.ListOpts{
		Pagination: sUtils.GetOrDefaultPagination(requestQuery.Pagination),
		Search:     requestQuery.Search,
		All:        requestQuery.All,
	})
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	usersDto := make([]body.UserRead, 0)
	for _, user := range userList {
		role := config.Config.GetRole(user.EffectiveRole.Name)
		usage, _ := deployV2.Users().GetUsage(user.ID)
		usersDto = append(usersDto, user.ToDTO(role, usage, deployV2.SMs().GetUrlByUserID(user.ID)))
	}

	context.Ok(usersDto)
}

// UpdateUser
// @Summary Update user
// @Description Update user
// @Tags User
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param userId path string true "User ID"
// @Param body body body.UserUpdate true "User update"
// @Success 200 {object} body.UserRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v2/users/{userId} [post]
func UpdateUser(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.UserUpdate
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	var userUpdate body.UserUpdate
	if err := context.GinContext.ShouldBindJSON(&userUpdate); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, AuthInfoNotAvailableErr)
		return
	}

	if requestURI.UserID == "" {
		requestURI.UserID = auth.User.ID
	}

	var effectiveRole *model.Role

	deployV2 := service.V2(auth)

	updated, err := deployV2.Users().Update(requestURI.UserID, &userUpdate)
	if err != nil {
		switch {
		case errors.Is(err, sErrors.UserNotFoundErr):
			context.NotFound("User not found")
		}

		context.ServerError(err, InternalError)
		return
	}

	if updated == nil {
		context.NotFound("User not found")
		return
	}

	usage, err := deployV2.Users().GetUsage(updated.ID)
	context.JSONResponse(200, updated.ToDTO(effectiveRole, usage, deployV2.SMs().GetUrlByUserID(updated.ID)))
}
