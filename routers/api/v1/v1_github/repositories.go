package v1_github

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/query"
	"go-deploy/pkg/status_codes"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/deployment_service"
	"net/http"
)

func ListGitHubRepositories(c *gin.Context) {
	context := sys.NewContext(c)

	var requestBody query.GitHubRepositoriesList
	if err := context.GinContext.Bind(&requestBody); err != nil {
		context.JSONResponse(http.StatusBadRequest, v1.CreateBindingError(err))
		return
	}

	// fetch access token
	accessToken, err := deployment_service.GetGitHubAccessTokenByCode(requestBody.Code)
	if err != nil {
		context.ErrorResponse(http.StatusBadRequest, status_codes.Error, fmt.Sprintf("Failed to fetch access token. details: %s", err))
		return
	}

	valid, reason, err := deployment_service.ValidGitHubToken(accessToken)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to validate access token. details: %s", err))
		return
	}

	if !valid {
		context.ErrorResponse(http.StatusBadRequest, status_codes.Error, fmt.Sprintf("Failed to validate access token. reason: %s", reason))
		return
	}

	// fetch repositories
	repositories, err := deployment_service.GetGitHubRepositories(accessToken)
	if err != nil {
		context.ErrorResponse(http.StatusInternalServerError, status_codes.Error, fmt.Sprintf("Failed to fetch repositories. details: %s", err))
		return
	}

	repositoriesDTOs := make([]body.GitHubRepository, 0)
	for _, repository := range repositories {
		repositoriesDTOs = append(repositoriesDTOs, repository.ToDTO())
	}

	context.JSONResponse(http.StatusOK, body.GitHubRepositoriesRead{
		AccessToken:  accessToken,
		Repositories: repositoriesDTOs,
	})
}
