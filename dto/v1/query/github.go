package query

type GitHubRepositoriesList struct {
	Code string `form:"code" binding:"required,min=1,max=100"`
}
