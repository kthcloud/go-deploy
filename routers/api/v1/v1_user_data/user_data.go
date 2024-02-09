package v1_user_data

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/models/dto/v1/body"
	"go-deploy/models/dto/v1/query"
	"go-deploy/models/dto/v1/uri"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service"
	sErrors "go-deploy/service/errors"
	"go-deploy/service/v1/user_data/opts"
	sUtils "go-deploy/service/v1/utils"
)

// Get
// @Summary Get user data by id
// @Description Get user data by id
// @Tags User data
// @Accept  json
// @Produce  json
// @Param id path string true "User data ID"
// @Success 200 {object}  body.UserDataRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 404 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /userData/{id} [get]
func Get(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.UserDataGet
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	deployV1 := service.V1(auth)

	userData, err := deployV1.UserData().Get(requestURI.ID)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if userData == nil {
		context.NotFound("User data not found")
		return
	}

	context.Ok(userData.ToDTO())
}

// List
// @Summary Get userdata list
// @Description Get userdata list
// @Tags User data
// @Accept  json
// @Produce  json
// @Param all query bool false "Want all users"
// @Success 200 {array}  body.UserRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /userData [get]
func List(c *gin.Context) {
	context := sys.NewContext(c)

	var requestQuery query.UserDataList
	if err := context.GinContext.ShouldBind(&requestQuery); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	deployV1 := service.V1(auth)

	userDataList, err := deployV1.UserData().List(opts.ListOpts{
		Pagination: sUtils.GetOrDefaultPagination(requestQuery.Pagination),
		UserID:     requestQuery.UserID,
	})
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	dtoUserData := make([]body.UserDataRead, len(userDataList))
	for idx, userData := range userDataList {
		dtoUserData[idx] = userData.ToDTO()
	}

	context.Ok(dtoUserData)
}

// Create
// @Summary Create user data
// @Description Create user data
// @Tags User data
// @Accept  json
// @Produce  json
// @Param body body body.UserDataCreate true "User data create"
// @Success 200 {object} body.UserDataRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /userData [post]
func Create(c *gin.Context) {
	context := sys.NewContext(c)

	var userCreate body.UserDataCreate
	if err := context.GinContext.ShouldBindJSON(&userCreate); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	deployV1 := service.V1(auth)

	exists, err := deployV1.UserData().Exists(userCreate.ID)
	if err != nil {
		context.ServerError(err, v1.InternalError)
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

		context.ServerError(err, v1.InternalError)
		return
	}

	context.Ok(userData.ToDTO())
}

// Update
// @Summary Update user data by id, create if not exists
// @Description Update user data by id, create if not exists
// @Tags User
// @Accept  json
// @Produce  json
// @Param id path string true "User data ID"
// @Param body body body.UserDataUpdate true "User data update"
// @Success 200 {object} body.UserDataRead
// @Failure 400 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /usersData/{id} [post]
func Update(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.UserDataUpdate
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	var userUpdate body.UserDataUpdate
	if err := context.GinContext.ShouldBindJSON(&userUpdate); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	deployV1 := service.V1(auth)

	userData, err := deployV1.UserData().Update(requestURI.ID, userUpdate.Data)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if userData == nil {
		userData, err = deployV1.UserData().Create(requestURI.ID, userUpdate.Data, auth.UserID)
		if err != nil {
			context.ServerError(err, v1.InternalError)
			return
		}

		if userData == nil {
			context.ServerError(fmt.Errorf("failed to create user data (when creating from update)"), v1.InternalError)
			return
		}
	}

	context.Ok(userData.ToDTO())
}

// Delete
// @Summary Delete user data by id
// @Description Delete user data by id
// @Tags User data
// @Accept  json
// @Produce  json
// @Param id path string true "User data ID"
// @Success 204 "No Content"
// @Failure 400 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /userData/{id} [delete]
func Delete(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.UserDataGet
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ServerError(err, v1.AuthInfoNotAvailableErr)
		return
	}

	deployV1 := service.V1(auth)

	err = deployV1.UserData().Delete(requestURI.ID)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	context.OkNoContent()
}
