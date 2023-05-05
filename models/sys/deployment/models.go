package deployment

import (
	harborModels "go-deploy/pkg/subsystems/harbor/models"
	k8sModels "go-deploy/pkg/subsystems/k8s/models"
	npmModels "go-deploy/pkg/subsystems/npm/models"
)

const (
	ActivityBeingCreated = "beingCreated"
	ActivityBeingDeleted = "beingDeleted"
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

type Env struct {
	Name  string `json:"name" bson:"name"`
	Value string `json:"value" bson:"value"`
}

type GithubActionConfig struct {
	Name string `yaml:"name"`
	On   On     `yaml:"on"`
	Jobs Jobs   `yaml:"jobs"`
}

type Push struct {
	Branches []string `yaml:"branches"`
}

type On struct {
	Push Push `yaml:"push"`
}

type With struct {
	Registry string `yaml:"registry,omitempty"`
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
	Push     bool   `yaml:"push,omitempty"`
	Tags     string `yaml:"tags,omitempty"`
}

type Steps struct {
	Name string `yaml:"name"`
	Uses string `yaml:"uses"`
	With With   `yaml:"with,omitempty"`
}

type Docker struct {
	RunsOn string  `yaml:"runs-on"`
	Steps  []Steps `yaml:"steps"`
}

type Jobs struct {
	Docker Docker `yaml:"docker"`
}

type UpdateParams struct {
	Private *bool  `json:"private" bson:"private"`
	Envs    *[]Env `json:"envs" bson:"envs"`
}

type CreateParams struct {
	Name    string `json:"name" bson:"name"`
	Private bool   `json:"private" bson:"private"`
	Envs    []Env  `json:"envs" bson:"envs"`
}
