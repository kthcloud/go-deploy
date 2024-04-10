package v1

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/dto/v1/body"
	"go-deploy/dto/v1/query"
	"go-deploy/dto/v1/uri"
	"go-deploy/pkg/sys"
	"go-deploy/service"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/v1/user_data/opts"
	sUtils "go-deploy/service/v1/utils"
)

// GetUserData
// @Summary Get user data
// @Description Get user data
// @Tags UserData
// @Accept  json
// @Produce  json
// @Param id path string true "User data ID"
// @Success 200 {object}  body.UserDataRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v1/userData/{id} [get]
func GetUserData(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.UserDataGet
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

	userData, err := deployV1.UserData().Get(requestURI.ID)
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	if userData == nil {
		context.NotFound("User data not found")
		return
	}

	context.Ok(userData.ToDTO())
}

// ListUserData
// @Summary List user data
// @Description List user data
// @Tags UserData
// @Accept  json
// @Produce  json
// @Param all query bool false "Want all users"
// @Success 200 {array}  body.UserRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v1/userData [get]
func ListUserData(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.UserDataList
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

	userDataList, err := deployV1.UserData().List(opts.ListOpts{
		Pagination: sUtils.GetOrDefaultPagination(requestQuery.Pagination),
		UserID:     requestQuery.UserID,
	})
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	dtoUserData := make([]body.UserDataRead, len(userDataList))
	for idx, userData := range userDataList {
		dtoUserData[idx] = userData.ToDTO()
	}

	context.Ok(dtoUserData)
}

// CreateUserData
// @Summary Create user data
// @Description Create user data
// @Tags UserData
// @Accept  json
// @Produce  json
// @Param body body body.UserDataCreate true "User data create"
// @Success 200 {object} body.UserDataRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v1/userData [post]
func CreateUserData(c *gin.Context) {
	context := sys.NewContext(c)

	var userCreate body.UserDataCreate
	if err := context.GinContext.ShouldBindJSON(&userCreate); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, AuthInfoNotAvailableErr)
		return
	}

	deployV1 := service.V1(auth)

	exists, err := deployV1.UserData().Exists(userCreate.ID)
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	if exists {
		context.UserError("User data already exists")
		return
	}

	userData, err := deployV1.UserData().Create(userCreate.ID, userCreate.Data, auth.UserID)
	if err != nil {
		var quotaErr *sErrors.QuotaExceededError
		if errors.As(err, &quotaErr) {
			context.UserError(quotaErr.Error())
			return
		}

		context.ServerError(err, InternalError)
		return
	}

	context.Ok(userData.ToDTO())
}

// UpdateUserData
// @Summary Update user data, create if not exists
// @Description Update user data, create if not exists
// @Tags User
// @Accept  json
// @Produce  json
// @Param id path string true "User data ID"
// @Param body body body.UserDataUpdate true "User data update"
// @Success 200 {object} body.UserDataRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v1/usersData/{id} [post]
func UpdateUserData(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.UserDataUpdate
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	var userUpdate body.UserDataUpdate
	if err := context.GinContext.ShouldBindJSON(&userUpdate); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	auth, err := WithAuth(&context)
	if err != nil {
		context.ServerError(err, AuthInfoNotAvailableErr)
		return
	}

	deployV1 := service.V1(auth)

	userData, err := deployV1.UserData().Update(requestURI.ID, userUpdate.Data)
	if err != nil {
		context.ServerError(err, InternalError)
		return
	}

	if userData == nil {
		userData, err = deployV1.UserData().Create(requestURI.ID, userUpdate.Data, auth.UserID)
		if err != nil {
			context.ServerError(err, InternalError)
			return
		}

		if userData == nil {
			context.ServerError(fmt.Errorf("failed to create user data (when creating from update)"), InternalError)
			return
		}
	}

	context.Ok(userData.ToDTO())
}

// DeleteUserData
// @Summary Delete user data
// @Description Delete user data
// @Tags UserData
// @Accept  json
// @Produce  json
// @Param id path string true "User data ID"
// @Success 204 "No Content"
// @Failure 400 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v1/userData/{id} [delete]
func DeleteUserData(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.UserDataGet
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

	err = deployV1.UserData().Delete(requestURI.ID)
	if err != nil {
		if errors.Is(err, sErrors.UserDataNotFoundErr) {
			context.NotFound("User data not found")
			return
		}

		context.ServerError(err, InternalError)
		return
	}

	context.OkNoContent()
}
