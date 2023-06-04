package v1_github

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/uri"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/deployment_service"
	"net/http"
)

func ListGitHubRepositories(c *gin.Context) {
	context := sys.NewContext(c)

	var requestBody uri.GitHubRepositoriesList
	if err := context.GinContext.BindJSON(&requestBody); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	// fetch access token
	accessToken, err := deployment_service.GetGitHubAccessTokenByCode(requestBody.Code)
	if err != nil {
		context.JSONResponse(http.StatusBadRequest, fmt.Sprintf("Failed to fetch access token. details: %s", err))
	}

	context.JSONResponse(http.StatusOK, body.GitHubRepositoriesRead{
		AccessToken:  accessToken,
		Repositories: make([]body.GitHubRepository, 0),
	})
	return
}
