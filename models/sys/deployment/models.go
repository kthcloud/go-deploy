package deployment

import (
	githubModels "go-deploy/pkg/subsystems/github/models"
	harborModels "go-deploy/pkg/subsystems/harbor/models"
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
	"time"
)

type Subsystems struct {
	K8s    K8s    `bson:"k8s"`
	Harbor Harbor `bson:"harbor"`
	GitHub GitHub `bson:"github"`
	GitLab GitLab `bson:"gitlab"`
}

type K8s struct {
	Namespace  k8sModels.NamespacePublic  `bson:"namespace"`
	Deployment k8sModels.DeploymentPublic `bson:"deployment"`
	Service    k8sModels.ServicePublic    `bson:"service"`
	Ingress    k8sModels.IngressPublic    `bson:"ingress"`
}

type Harbor struct {
	Project    harborModels.ProjectPublic    `bson:"project"`
	Robot      harborModels.RobotPublic      `bson:"robot"`
	Repository harborModels.RepositoryPublic `bson:"repository"`
	Webhook    harborModels.WebhookPublic    `bson:"webhook"`
}

type GitHub struct {
	Webhook     githubModels.WebhookPublic `bson:"webhook"`
	Placeholder bool                       `bson:"placeholder"`
}

type GitLab struct {
	LastBuild GitLabBuild `bson:"lastBuild"`
}

type GitLabBuild struct {
	ID        int       `bson:"id"`
	ProjectID int       `bson:"projectId"`
	Trace     []string  `bson:"trace"`
	Status    string    `bson:"status"`
	Stage     string    `bson:"stage"`
	CreatedAt time.Time `bson:"createdAt"`
}

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
	ZoneID  string              `json:"zoneId,omitempty" bson:"zoneId,omitempty"`
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
