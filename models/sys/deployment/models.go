package deployment

type Env struct {
	Name  string `json:"name" bson:"name"`
	Value string `json:"value" bson:"value"`
}

type Usage struct {
	Count int `json:"deployments"`
}

type UpdateParams struct {
	Private      *bool     `json:"private" bson:"private"`
	Envs         *[]Env    `json:"envs" bson:"envs"`
	ExtraDomains *[]string `json:"extraDomains" bson:"extraDomains"`
}

type GitHubCreateParams struct {
	Token        string `json:"token" bson:"token"`
	RepositoryID int64  `json:"repositoryId" bson:"repositoryId"`
}

type CreateParams struct {
	Name    string              `json:"name" bson:"name"`
	Private bool                `json:"private" bson:"private"`
	Envs    []Env               `json:"envs" bson:"envs"`
	GitHub  *GitHubCreateParams `json:"github,omitempty" bson:"github,omitempty"`
	Zone    string              `json:"zone,omitempty" bson:"zoneId,omitempty"`
}

type BuildParams struct {
	Tag       string `json:"tag" bson:"tag"`
	Branch    string `json:"branch" bson:"branch"`
	ImportURL string `json:"importUrl" bson:"importUrl"`
}

type GitHubRepository struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Owner         string `json:"owner"`
	CloneURL      string `json:"clone_url"`
	DefaultBranch string `json:"default_branch"`
}

type GitHubWebhook struct {
	ID     int64    `json:"id"`
	Events []string `json:"events"`
}
