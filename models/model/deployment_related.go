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

	// VisibilityPublic is a public app.
	VisibilityPublic = "public"
	// VisibilityPrivate is a private app.
	VisibilityPrivate = "private"
	// VisibilityAuth is an app that requires authentication.
	VisibilityAuth = "auth"
)

var (
	EmptyReplicaStatus = &ReplicaStatus{}
)

type App struct {
	Name string `bson:"name"`

	CpuCores float64         `bson:"cpuCores,omitempty"`
	RAM      float64         `bson:"ram,omitempty"`
	Replicas int             `bson:"replicas"`
	GPUs     []DeploymentGPU `bson:"gpus,omitempty"`

	Image         string             `bson:"image"`
	InternalPort  int                `bson:"internalPort"`
	InternalPorts []int              `bson:"internalPorts"`
	Envs          []DeploymentEnv    `bson:"envs"`
	Volumes       []DeploymentVolume `bson:"volumes"`
	Visibility    string             `bson:"visibility"`

	// Deprecated: use Visibility instead.
	Private bool `bson:"private"`

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

type PodDeleted struct {
	DeploymentName string `bson:"deploymentName"`
	PodName        string `bson:"podName"`
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

type DeploymentGPU struct {
	Name         string `bson:"name"`
	TemplateName string `bson:"name"`
}

type DeploymentUsage struct {
	CpuCores float64
	RAM      float64
}

type DeploymentError struct {
	Reason      string `bson:"reason"`
	Description string `bson:"description"`
}

type DeploymentStatus struct {
	Name                string `bson:"name"`
	Generation          int    `bson:"generation"`
	DesiredReplicas     int    `bson:"desiredReplicas"`
	ReadyReplicas       int    `bson:"readyReplicas"`
	AvailableReplicas   int    `bson:"availableReplicas"`
	UnavailableReplicas int    `bson:"unavailableReplicas"`
}

type DeploymentEvent struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Reason      string `json:"reason"`
	Description string `json:"description"`
	ObjectKind  string `json:"objectKind"`
}
