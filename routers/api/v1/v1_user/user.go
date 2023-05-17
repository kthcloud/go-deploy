package v1_user

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/query"
	"go-deploy/models/dto/uri"
	userModel "go-deploy/models/sys/user"
	"go-deploy/pkg/app"
	"go-deploy/pkg/status_codes"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/user_service"
	"net/http"
)

// GetList
// @Summary Get user list
// @Description Get user list
// @Tags User
// @Accept  json
// @Produce  json
// @Param wantAll query bool false "Want all users"
// @Success 200 {array}  body.UserRead
// @Failure 400 {object} app.ErrorResponse
// @Failure 500 {object} app.ErrorResponse
// @Router /api/v1/users [get]
func GetList(c *gin.Context) {
	context := app.NewContext(c)

	var requestQuery query.UserList
	if err := context.GinContext.Bind(&requestQuery); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	if requestQuery.WantAll && auth.IsAdmin() {
		users, err := user_service.GetAll()
		if err != nil {
			context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
			return
		}

		usersDto := make([]body.UserRead, 0)
		for _, user := range users {
			if user.ID == auth.UserID {
				updatedUser, err := user_service.GetOrCreate(auth.UserID, auth.JwtToken.PreferredUsername, auth.Roles)
				if err != nil {
					continue
				}

				if updatedUser != nil {
					user = *updatedUser
				}
			}

			quota, err := user_service.GetQuotaByUserID(user.ID)
			if err != nil {
				quota = &userModel.Quota{}
			}

			usersDto = append(usersDto, user.ToDTO(quota))
		}

		context.JSONResponse(200, usersDto)
		return
	}

	user, err := user_service.GetByID(auth.UserID, auth.UserID, auth.IsAdmin())
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	if user == nil {
		context.ErrorResponse(http.StatusNotFound, status_codes.ResourceNotFound, fmt.Sprintf("User with id %s not found", auth.UserID))
		return
	}

	quota, err := user_service.GetQuotaByUserID(user.ID)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to update user: %s", err))
		return
	}

	context.JSONResponse(200, user.ToDTO(quota))
}

// Get
// @Summary Get user by id
// @Description Get user by id
// @Tags User
// @Accept  json
// @Produce  json
// @Param userId path string true "User ID"
// @Success 200 {object}  body.UserRead
// @Failure 400 {object} app.ErrorResponse
// @Failure 500 {object} app.ErrorResponse
// @Router /api/v1/users/{userId} [get]
func Get(c *gin.Context) {
	context := app.NewContext(c)

	var requestURI uri.UserGet
	if err := context.GinContext.BindUri(&requestURI); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	requestedUserID := requestURI.UserID
	if requestedUserID == "" {
		requestedUserID = auth.UserID
	}

	var user *userModel.User
	if requestedUserID == auth.UserID {
		user, err = user_service.GetOrCreate(auth.UserID, auth.JwtToken.PreferredUsername, auth.Roles)
		if err != nil {
			context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get user: %s", err))
			return
		}
	} else {
		user, err = user_service.GetByID(requestedUserID, auth.UserID, auth.IsAdmin())
		if err != nil {
			context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get user: %s", err))
			return
		}

		if user == nil {
			context.ErrorResponse(http.StatusNotFound, status_codes.ResourceNotFound, fmt.Sprintf("User with id %s not found", requestedUserID))
			return
		}
	}

	quota, err := user_service.GetQuotaByUserID(user.ID)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to update user: %s", err))
		return
	}

	context.JSONResponse(200, user.ToDTO(quota))
}

func Update(c *gin.Context) {
	context := app.NewContext(c)

	var requestURI uri.UserUpdate
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	var userUpdate body.UserUpdate
	if err := context.GinContext.ShouldBindJSON(&userUpdate); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	requestedUserID := requestURI.UserID
	if requestedUserID == "" {
		requestedUserID = auth.UserID
	}

	user, err := user_service.GetByID(requestedUserID, auth.UserID, auth.IsAdmin())
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to update user: %s", err))
		return
	}

	if user == nil {
		context.ErrorResponse(http.StatusNotFound, status_codes.ResourceNotFound, fmt.Sprintf("User with id %s not found", requestedUserID))
		return
	}

	err = user_service.Update(requestedUserID, auth.UserID, auth.IsAdmin(), &userUpdate)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to update user: %s", err))
		return
	}

	updatedUser, err := user_service.GetByID(requestedUserID, auth.UserID, auth.IsAdmin())
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to update user: %s", err))
		return
	}

	if updatedUser == nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("User with id %s not found", requestedUserID))
		return
	}

	quota, err := user_service.GetQuotaByUserID(updatedUser.ID)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to update user: %s", err))
		return
	}

	context.JSONResponse(200, updatedUser.ToDTO(quota))
}
