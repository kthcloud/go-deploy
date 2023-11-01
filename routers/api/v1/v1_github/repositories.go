package v1_github

import (
	"github.com/gin-gonic/gin"
	"go-deploy/models/dto/body"
	"go-deploy/models/dto/query"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service/deployment_service"
)

// ListGitHubRepositories
// Get
// @Summary Get GitHub repositories
// @Tags GitHub
// @Accept  json
// @Produce  json
// @Param code query string true "code"
// @Success 200 {object} body.GitHubRepositoriesRead
// @Router /github/repositories [get]
func ListGitHubRepositories(c *gin.Context) {
	context := sys.NewContext(c)

	var requestBody query.GitHubRepositoriesList
	if err := context.GinContext.Bind(&requestBody); err != nil {
		context.BindingError(v1.CreateBindingError(err))
		return
	}

	// fetch access token
	accessToken, err := deployment_service.GetGitHubAccessTokenByCode(requestBody.Code)
	if err != nil {
		context.Unauthorized("Failed to get GitHub access token from code")
		return
	}

	valid, reason, err := deployment_service.ValidGitHubToken(accessToken)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if !valid {
		context.Unauthorized(reason)
		return
	}

	// fetch repositories
	repositories, err := deployment_service.GetGitHubRepositories(accessToken)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	repositoriesDTOs := make([]body.GitHubRepository, 0)
	for _, repository := range repositories {
		repositoriesDTOs = append(repositoriesDTOs, repository.ToDTO())
	}

	context.Ok(body.GitHubRepositoriesRead{
		AccessToken:  accessToken,
		Repositories: repositoriesDTOs,
	})
}
