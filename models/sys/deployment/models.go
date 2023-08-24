package deployment

import "go-deploy/models/sys/deployment/subsystems"

type App struct {
	Name         string   `bson:"name"`
	Private      bool     `bson:"private"`
	Envs         []Env    `bson:"envs"`
	Volumes      []Volume `bson:"volumes"`
	InitCommands []string `bson:"initCommands"`
	ExtraDomains []string `bson:"extraDomains"`
}

type Subsystems struct {
	K8s    subsystems.K8s    `bson:"k8s"`
	Harbor subsystems.Harbor `bson:"harbor"`
	GitHub subsystems.GitHub `bson:"github"`
	GitLab subsystems.GitLab `bson:"gitlab"`
}

type Env struct {
	Name  string `json:"name" bson:"name"`
	Value string `json:"value" bson:"value"`
}

type Volume struct {
	Name       string `bson:"name"`
	Init       bool   `bson:"init"`
	AppPath    string `bson:"appPath"`
	ServerPath string `bson:"serverPath"`
}

type Usage struct {
	Count int `json:"deployments"`
}

type UpdateParams struct {
	Private      *bool     `json:"private" bson:"private"`
	Envs         *[]Env    `json:"envs" bson:"envs"`
	Volumes      *[]Volume `json:"volumes" bson:"volumes"`
	InitCommands *[]string `json:"initCommands" bson:"initCommands"`
	ExtraDomains *[]string `json:"extraDomains" bson:"extraDomains"`
}

type GitHubCreateParams struct {
	Token        string `json:"token" bson:"token"`
	RepositoryID int64  `json:"repositoryId" bson:"repositoryId"`
}

type CreateParams struct {
	Name         string              `json:"name" bson:"name"`
	Private      bool                `json:"private" bson:"private"`
	Envs         []Env               `json:"envs" bson:"envs"`
	Volumes      []Volume            `json:"volumes" bson:"volumes"`
	InitCommands []string            `json:"initCommands" bson:"initCommands"`
	GitHub       *GitHubCreateParams `json:"github,omitempty" bson:"github,omitempty"`
	Zone         string              `json:"zone,omitempty" bson:"zoneId,omitempty"`
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
