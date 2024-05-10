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

	// ReplicaStatus is a group of fields that describe the status of the replicas.
	// It is only set for apps that has status update.
	ReplicaStatus *ReplicaStatus `bson:"replicaStatus,omitempty"`

	PingPath   string `bson:"pingPath"`
	PingResult int    `bson:"pingResult"`
}

type ReplicaStatus struct {
	// DesiredReplicas is the number of replicas that the deployment should have.
	DesiredReplicas int `bson:"desiredReplicas"`
	// ReadyReplicas is the number of replicas that are ready.
	ReadyReplicas int `bson:"readyReplicas"`
	// AvailableReplicas is the number of replicas that are available.
	AvailableReplicas int `bson:"availableReplicas"`
	// UnavailableReplicas is the number of replicas that are unavailable.
	UnavailableReplicas int `bson:"unavailableReplicas"`
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

type DeploymentUsage struct {
	Replicas int
}

type DeploymentError struct {
	Reason      string `bson:"reason"`
	Description string `bson:"description"`
}
