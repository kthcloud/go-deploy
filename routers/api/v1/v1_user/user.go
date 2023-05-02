package v1_user

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/pkg/app"
	"go-deploy/pkg/status_codes"
	"go-deploy/pkg/validator"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/user_service"
	"net/http"
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
	wantAll := context.GinContext.Query("all") == "true"
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
