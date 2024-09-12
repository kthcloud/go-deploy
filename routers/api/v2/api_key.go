package v2

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/kthcloud/go-deploy/dto/v2/body"
	"github.com/kthcloud/go-deploy/dto/v2/uri"
	"github.com/kthcloud/go-deploy/pkg/sys"
	"github.com/kthcloud/go-deploy/service"
	sErrors "github.com/kthcloud/go-deploy/service/errors"
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
// @Router /v2/users/{userId}/apiKeys [post]
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

	deployV2 := service.V2(auth)

	apiKey, err := deployV2.Users().ApiKeys().Create(requestURI.UserID, &requestBody)
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
