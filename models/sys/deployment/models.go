package deployment

import (
	"go-deploy/models/sys/deployment/subsystems"
	"time"
)

const (
	// TypeCustom is a deployment that builds its own image, e.g., with a Dockerfile.
	TypeCustom = "custom"
	// TypePrebuilt is a deployment that uses a prebuilt image, such as nginx:latest.
	TypePrebuilt = "prebuilt"

	// LogSourcePod is a log source for a pod in Kubernetes.
	LogSourcePod = "pod"
	// LogSourceDeployment is a log source for a deployment in go-deploy.
	LogSourceDeployment = "deployment"
	// LogSourceBuild is a log source for a build in GitLab CI.
	LogSourceBuild = "build"

	// CustomDomainStatusPending is the status of a custom domain that is pending verification.
	CustomDomainStatusPending = "pending"
	// CustomDomainStatusVerificationFailed is the status of a custom domain that failed verification.
	// This is either caused by the DNS record not being set or the DNS record not being propagated yet.
	CustomDomainStatusVerificationFailed = "verificationFailed"
	// CustomDomainStatusActive is the status of a custom domain that is active,
	// i.e., the DNS record is set and propagated.
	CustomDomainStatusActive = "active"
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

	Args         []string `bson:"args"`
	InitCommands []string `bson:"initCommands"`

	CustomDomain *CustomDomain `bson:"customDomain"`

	PingPath   string `bson:"pingPath"`
	PingResult int    `bson:"pingResult"`
}

type Subsystems struct {
	K8s    subsystems.K8s    `bson:"k8s"`
	Harbor subsystems.Harbor `bson:"harbor"`
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
