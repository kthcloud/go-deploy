package deployment

import (
	"go-deploy/models/sys/deployment/subsystems"
	"time"
)

const (
	TypeCustom   = "custom"
	TypePrebuilt = "prebuilt"

	LogSourcePod        = "pod"
	LogSourceDeployment = "deployment"
	LogSourceBuild      = "build"

	CustomDomainStatusPending            = "pending"
	CustomDomainStatusVerificationFailed = "verificationFailed"
	CustomDomainStatusActive             = "active"
)

type CustomDomain struct {
	Domain string `bson:"domain"`
	Secret string `bson:"secret"`
	Status string `bson:"status"`
}

type App struct {
	Name string `bson:"name"`

	Image        string   `bson:"image"`
	InternalPort int      `bson:"internalPort"`
	Private      bool     `bson:"private"`
	Replicas     int      `bson:"replicas"`
	Envs         []Env    `bson:"envs"`
	Volumes      []Volume `bson:"volumes"`
	InitCommands []string `bson:"initCommands"`

	CustomDomain *CustomDomain `bson:"customDomain"`

	PingPath   string `bson:"pingPath"`
	PingResult int    `bson:"pingResult"`
}

type Subsystems struct {
	K8s    subsystems.K8s    `bson:"k8s"`
	Harbor subsystems.Harbor `bson:"harbor"`
	GitHub subsystems.GitHub `bson:"github"`
	GitLab subsystems.GitLab `bson:"gitlab"`
}

type Log struct {
	Source    string    `bson:"source"`
	Prefix    string    `bson:"prefix"`
	Line      string    `bson:"line"`
	CreatedAt time.Time `bson:"createdAt"`
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

type Transfer struct {
	Code   string `bson:"code"`
	UserID string `bson:"userId"`
}

type Usage struct {
	Count int
}

type GitHubRepository struct {
	ID            int64
	Name          string
	Owner         string
	CloneURL      string
	DefaultBranch string
}

type GitHubWebhook struct {
	ID     int64
	Events []string
}
