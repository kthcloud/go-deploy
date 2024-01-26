package v1_github

import (
	"github.com/gin-gonic/gin"
	"go-deploy/models/dto/v1/body"
	"go-deploy/models/dto/v1/query"
	"go-deploy/pkg/sys"
	v1 "go-deploy/routers/api/v1"
	"go-deploy/service"
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

	deployV1 := service.V1()

	// Fetch access token
	accessToken, err := deployV1.Deployments().GetGitHubAccessTokenByCode(requestBody.Code)
	if err != nil {
		context.Unauthorized("Failed to get GitHub access token from code")
		return
	}

	valid, reason, err := deployV1.Deployments().ValidGitHubToken(accessToken)
	if err != nil {
		context.ServerError(err, v1.InternalError)
		return
	}

	if !valid {
		context.Unauthorized(reason)
		return
	}

	// Fetch repositories
	repositories, err := deployV1.Deployments().GetGitHubRepositories(accessToken)
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
