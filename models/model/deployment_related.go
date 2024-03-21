package model

import (
	"time"
)

const (
	// DeploymentTypeCustom is a deployment that builds its own image, e.g., with a Dockerfile.
	DeploymentTypeCustom = "custom"
	// DeploymentTypePrebuilt is a deployment that uses a prebuilt image, such as nginx:latest.
	DeploymentTypePrebuilt = "prebuilt"

	// LogSourcePod is a log source for a pod in Kubernetes.
	LogSourcePod = "pod"
	// LogSourceDeployment is a log source for a deployment in go-deploy.
	LogSourceDeployment = "deployment"
	// LogSourceBuild is a log source for a build in GitLab CI.
	LogSourceBuild = "build"
)

type App struct {
	Name string `bson:"name"`

	Image        string             `bson:"image"`
	InternalPort int                `bson:"internalPort"`
	Private      bool               `bson:"private"`
	Replicas     int                `bson:"replicas"`
	Envs         []DeploymentEnv    `bson:"envs"`
	Volumes      []DeploymentVolume `bson:"volumes"`

	Args         []string `bson:"args"`
	InitCommands []string `bson:"initCommands"`

	CustomDomain *CustomDomain `bson:"customDomain"`

	PingPath   string `bson:"pingPath"`
	PingResult int    `bson:"pingResult"`
}

type Log struct {
	Source    string    `bson:"source"`
	Prefix    string    `bson:"prefix"`
	Line      string    `bson:"line"`
	CreatedAt time.Time `bson:"createdAt"`
}

type DeploymentEnv struct {
	Name  string `json:"name" bson:"name"`
	Value string `json:"value" bson:"value"`
}

type DeploymentVolume struct {
	Name       string `bson:"name"`
	Init       bool   `bson:"init"`
	AppPath    string `bson:"appPath"`
	ServerPath string `bson:"serverPath"`
}

type DeploymentTransfer struct {
	Code   string `bson:"code"`
	UserID string `bson:"userId"`
}

type DeploymentUsage struct {
	Count int
}
