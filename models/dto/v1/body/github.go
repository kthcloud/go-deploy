package body

type GitHubRepositoriesRead struct {
	AccessToken  string             `json:"accessToken"`
	Repositories []GitHubRepository `json:"repositories"`
}

type GitHubRepository struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}
