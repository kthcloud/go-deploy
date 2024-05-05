package v1

import (
	"errors"
	"github.com/gin-gonic/gin"
	"go-deploy/dto/v1/body"
	"go-deploy/dto/v1/uri"
	"go-deploy/pkg/sys"
	"go-deploy/service"
	sErrors "go-deploy/service/errors"
)

// CreateApiKey
// @Summary Create API key
// @Description Create API key
// @Tags User
// @Accept  json
// @Produce  json
// @Security ApiKeyAuth
// @Param userId path string true "User ID"
// @Param body body body.ApiKeyCreate true "API key create body"
// @Success 200 {object}  body.ApiKeyCreated
// @Failure 400 {object} sys.ErrorResponse
// @Failure 500 {object} sys.ErrorResponse
// @Router /v1/users/{userId}/apiKeys [post]
func CreateApiKey(c *gin.Context) {
	context := sys.NewContext(c)

	var requestURI uri.ApiKeyCreate
	if err := context.GinContext.ShouldBindUri(&requestURI); err != nil {
		context.BindingError(CreateBindingError(err))
		return
	}

	var requestBody body.ApiKeyCreate
	if err := context.GinContext.ShouldBindJSON(&requestBody); err != nil {
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

	deployV1 := service.V1(auth)

	apiKey, err := deployV1.Users().ApiKeys().Create(requestURI.UserID, &requestBody)
	if err != nil {
		switch {
		case errors.Is(err, sErrors.UserNotFoundErr):
			context.NotFound("User not found")
		case errors.Is(err, sErrors.ApiKeyNameTakenErr):
			context.UserError("API key name already taken")
		default:
			context.ServerError(err, InternalError)
		}
		return
	}

	context.Ok(apiKey.ToDTO())
}
