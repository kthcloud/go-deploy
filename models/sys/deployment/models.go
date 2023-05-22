package deployment

import (
	githubModels "go-deploy/pkg/subsystems/github/models"
	harborModels "go-deploy/pkg/subsystems/harbor/models"
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
	npmModels "go-deploy/pkg/subsystems/npm/models"
)

const (
	ActivityBeingCreated = "beingCreated"
	ActivityBeingDeleted = "beingDeleted"
	ActivityRestarting   = "restarting"
	ActivityBuilding     = "building"
)

type Deployment struct {
	ID      string `bson:"id"`
	Name    string `bson:"name"`
	OwnerID string `bson:"ownerId"`

	Private bool  `bson:"private"`
	Envs    []Env `bson:"envs"`

	Activities []string `bson:"activities"`

	Subsystems    Subsystems `bson:"subsystems"`
	StatusCode    int        `bson:"statusCode"`
	StatusMessage string     `bson:"statusMessage"`
}

type Subsystems struct {
	K8s    K8s    `bson:"k8s"`
	Npm    NPM    `bson:"npm"`
	Harbor Harbor `bson:"harbor"`
	GitHub GitHub `bson:"github"`
}

type K8s struct {
	Namespace  k8sModels.NamespacePublic  `bson:"namespace"`
	Deployment k8sModels.DeploymentPublic `bson:"deployment"`
	Service    k8sModels.ServicePublic    `bson:"service"`
}

type NPM struct {
	ProxyHost npmModels.ProxyHostPublic `bson:"proxyHost"`
}

type Harbor struct {
	Project    harborModels.ProjectPublic    `bson:"project"`
	Robot      harborModels.RobotPublic      `bson:"robot"`
	Repository harborModels.RepositoryPublic `bson:"repository"`
	Webhook    harborModels.WebhookPublic    `bson:"webhook"`
}

type GitHub struct {
	Webhook githubModels.WebhookPublic `bson:"webhook"`
}

type Env struct {
	Name  string `json:"name" bson:"name"`
	Value string `json:"value" bson:"value"`
}

type Usage struct {
	Count int `json:"deployments"`
}

type UpdateParams struct {
	Private *bool  `json:"private" bson:"private"`
	Envs    *[]Env `json:"envs" bson:"envs"`
}

type GitHubCreateParams struct {
	Token        string `json:"token" bson:"token"`
	RepositoryID int64  `json:"repositoryId" bson:"repositoryId"`
}

type CreateParams struct {
	Name    string              `json:"name" bson:"name"`
	Private bool                `json:"private" bson:"private"`
	Envs    []Env               `json:"envs" bson:"envs"`
	GitHub  *GitHubCreateParams `json:"omitempty,github" bson:"omitempty,github"`
}

type BuildParams struct {
	Tag       string `json:"tag" bson:"tag"`
	Branch    string `json:"branch" bson:"branch"`
	ImportURL string `json:"importUrl" bson:"importUrl"`
}
