package v1_user

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/models/dto"
	"go-deploy/pkg/app"
	"go-deploy/pkg/status_codes"
	"go-deploy/pkg/validator"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/user_service"
	"net/http"
	"strconv"
)

func GetList(c *gin.Context) {
	context := app.NewContext(c)

	rules := validator.MapData{
		"all": []string{"bool"},
	}

	validationErrors := context.ValidateQueryParams(&rules)
	if len(validationErrors) > 0 {
		context.ResponseValidationError(validationErrors)
		return
	}

	isAdmin := v1.IsAdmin(&context)
	wantAll, _ := strconv.ParseBool(context.GinContext.Query("all"))
	if isAdmin && wantAll {
		users, err := user_service.GetAll()
		if err != nil {
			context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
			return
		}

		context.JSONResponse(200, users)
		return
	}

	token, err := context.GetKeycloakToken()
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}
	userID := token.Sub
	isAdmin = v1.IsAdmin(&context)

	user, err := user_service.GetByID(userID, userID, isAdmin)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	if user == nil {
		context.ErrorResponse(http.StatusNotFound, status_codes.Error, fmt.Sprintf("User with id %s not found", userID))
		return
	}

	context.JSONResponse(200, user.ToDTO())
}

func Get(c *gin.Context) {
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
	}
	userID := token.Sub
	isAdmin := v1.IsAdmin(&context)

	requestedUserID := context.GinContext.Param("userId")
	if requestedUserID == "" {
		requestedUserID = userID
	}

	user, err := user_service.GetByID(requestedUserID, userID, isAdmin)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	if user == nil {
		if requestedUserID == userID {
			err = user_service.CreateUser(userID, token.PreferredUsername)
			if err != nil {
				context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to create user: %s", err))
				return
			}

			createdUser, err := user_service.GetByID(userID, userID, isAdmin)
			if err != nil {
				context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
				return
			}

			if createdUser == nil {
				context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to create user: %s", err))
				return
			}

			user = createdUser
		}
	}

	context.JSONResponse(200, user.ToDTO())
}

func Update(c *gin.Context) {
	context := app.NewContext(c)

	var params dto.UserUpdateParams
	err := context.GinContext.ShouldBindUri(&params)
	if err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(&params, err))
		return
	}

	var userUpdate dto.UserUpdate
	err = context.GinContext.ShouldBindJSON(&userUpdate)
	if err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(&userUpdate, err))
		return
	}

	// check if valid public key
	if userUpdate.PublicKeys != nil {
		for i, publicKey := range *userUpdate.PublicKeys {
			if !v1.IsValidSshPublicKey(publicKey) {
				bindingError := v1.CreateBindingErrorFromString("publicKeys", fmt.Sprintf("publicKeys[%d] is not a valid ssh public key", i))
				context.JSONResponse(http.StatusBadRequest, bindingError)
				return
			}
		}
	}

	token, err := context.GetKeycloakToken()
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
	}
	userID := token.Sub
	isAdmin := v1.IsAdmin(&context)

	requestedUserID := context.GinContext.Param("userId")
	if requestedUserID == "" {
		requestedUserID = userID
	}

	user, err := user_service.GetByID(requestedUserID, userID, isAdmin)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to update user: %s", err))
		return
	}

	if user == nil {
		context.ErrorResponse(http.StatusNotFound, status_codes.Error, fmt.Sprintf("User with id %s not found", requestedUserID))
		return
	}

	if !isAdmin && user.ID != userID {
		context.ErrorResponse(http.StatusForbidden, status_codes.Error, fmt.Sprintf("You don't have permission to update this user"))
		return
	}

	err = user_service.Update(requestedUserID, userID, isAdmin, userUpdate)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to update user: %s", err))
		return
	}

	updatedUser, err := user_service.GetByID(requestedUserID, userID, isAdmin)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to update user: %s", err))
		return
	}

	if updatedUser == nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("User with id %s not found", requestedUserID))
		return
	}

	context.JSONResponse(200, updatedUser.ToDTO())
}
