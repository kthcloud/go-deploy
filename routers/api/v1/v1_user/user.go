package v1_user

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/query"
	"go-deploy/models/dto/uri"
	"go-deploy/pkg/app"
	"go-deploy/pkg/status_codes"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/user_service"
	"net/http"
)

func GetList(c *gin.Context) {
	context := app.NewContext(c)

	var requestQuery query.UserList
	if err := context.GinContext.BindQuery(&requestQuery); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(&requestQuery, err))
		return
	}

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	if requestQuery.WantAll && auth.IsAdmin {
		users, err := user_service.GetAll()
		if err != nil {
			context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
			return
		}

		context.JSONResponse(200, users)
		return
	}

	user, err := user_service.GetByID(auth.UserID, auth.UserID, auth.IsAdmin)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	if user == nil {
		context.ErrorResponse(http.StatusNotFound, status_codes.ResourceNotFound, fmt.Sprintf("User with id %s not found", auth.UserID))
		return
	}

	context.JSONResponse(200, user.ToDTO())
}

func Get(c *gin.Context) {
	context := app.NewContext(c)

	var requestURI uri.UserGet
	if err := context.GinContext.BindUri(&requestURI); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(&requestURI, err))
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

	user, err := user_service.GetByID(requestedUserID, auth.UserID, auth.IsAdmin)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("%s", err))
		return
	}

	if user == nil {
		if requestedUserID == auth.UserID {
			err = user_service.CreateUser(auth.UserID, auth.JwtToken.PreferredUsername)
			if err != nil {
				context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to create user: %s", err))
				return
			}

			createdUser, err := user_service.GetByID(auth.UserID, auth.UserID, auth.IsAdmin)
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

	var requestURI uri.UserUpdate
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(&requestURI, err))
		return
	}

	var userUpdate body.UserUpdate
	if err := context.GinContext.ShouldBindJSON(&userUpdate); err != nil {
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

	auth, err := v1.WithAuth(&context)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to get auth info: %s", err))
		return
	}

	requestedUserID := requestURI.UserID
	if requestedUserID == "" {
		requestedUserID = auth.UserID
	}

	user, err := user_service.GetByID(requestedUserID, auth.UserID, auth.IsAdmin)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to update user: %s", err))
		return
	}

	if user == nil {
		context.ErrorResponse(http.StatusNotFound, status_codes.ResourceNotFound, fmt.Sprintf("User with id %s not found", requestedUserID))
		return
	}

	err = user_service.Update(requestedUserID, auth.UserID, auth.IsAdmin, userUpdate)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to update user: %s", err))
		return
	}

	updatedUser, err := user_service.GetByID(requestedUserID, auth.UserID, auth.IsAdmin)
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
