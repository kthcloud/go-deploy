package body

import (
	"time"
)

type DeploymentRead struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Type    string `json:"type"`
	OwnerID string `json:"ownerId"`
	Zone    string `json:"zone"`

	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty"`
	RepairedAt  *time.Time `json:"repairedAt,omitempty"`
	RestartedAt *time.Time `json:"restartedAt,omitempty"`
	AccessedAt  time.Time  `json:"accessedAt"`

	URL   *string         `json:"url,omitempty"`
	Specs DeploymentSpecs `json:"specs"`

	Envs            []Env             `json:"envs"`
	Volumes         []Volume          `json:"volumes"`
	InitCommands    []string          `json:"initCommands"`
	Args            []string          `json:"args"`
	InternalPort    int               `json:"internalPort"`
	InternalPorts   []int             `json:"internalPorts,omitempty"`
	Image           *string           `json:"image,omitempty"`
	HealthCheckPath *string           `json:"healthCheckPath,omitempty"`
	CustomDomain    *CustomDomainRead `json:"customDomain,omitempty"`
	Visibility      string            `json:"visibility"`

	NeverStale bool `json:"neverStale" bson:"neverStale" binding:"omitempty,boolean"`

	// Deprecated: Use Visibility instead.
	Private bool `json:"private"`

	Status        string         `json:"status"`
	Error         *string        `json:"error,omitempty"`
	ReplicaStatus *ReplicaStatus `json:"replicaStatus,omitempty"`
	PingResult    *int           `json:"pingResult,omitempty"`

	// Integrations are currently not used, but could be used if we wanted to add a list of integrations to the deployment
	//
	// For example GitHub
	Integrations []string `json:"integrations"`
	Teams        []string `json:"teams"`

	StorageURL *string `json:"storageUrl,omitempty"`
}

type DeploymentCreate struct {
	Name string `json:"name" bson:"name" binding:"required,rfc1035,min=3,max=30,deployment_name"`

	CpuCores *float64        `json:"cpuCores,omitempty" bson:"cpuCores,omitempty" binding:"omitempty,min=0.1"`
	RAM      *float64        `json:"ram,omitempty" bson:"ram,omitempty" binding:"omitempty,min=0.1"`
	Replicas *int            `json:"replicas,omitempty" bson:"replicas,omitempty" binding:"omitempty,min=0,max=100"`
	GPUs     []DeploymentGPU `json:"gpus,omitempty" bson:"gpus,omitempty" binding:"omitempty,min=0"`

	Envs         []Env    `json:"envs" bson:"envs" binding:"omitempty,env_list,min=0,max=1000,dive"`
	Volumes      []Volume `json:"volumes" bson:"volumes" binding:"omitempty,min=0,max=100,dive"`
	InitCommands []string `json:"initCommands" bson:"initCommands" binding:"omitempty,min=0,max=100,dive,min=0,max=100"`
	Args         []string `json:"args" bson:"args" binding:"omitempty,min=0,max=100,dive,min=0,max=100"`
	Visibility   string   `json:"visibility" bson:"visibility" binding:"omitempty,oneof=public private auth"`

	// Boolean to make deployment never get disabled, despite being stale
	NeverStale bool `json:"neverStale" bson:"neverStale" binding:"omitempty,boolean"`

	// Deprecated: Use Visibility instead.
	Private bool `json:"private" bson:"private" binding:"omitempty,boolean"`

	Image           *string `json:"image,omitempty" bson:"image,omitempty" binding:"omitempty,min=1,max=1000"`
	HealthCheckPath *string `json:"healthCheckPath" bson:"healthCheckPath,omitempty" binding:"omitempty,min=0,max=1000,health_check_path"`
	// CustomDomain is the domain that the deployment will be available on.
	// The max length is set to 243 to allow for a subdomain when confirming the domain.
	CustomDomain *string `json:"customDomain,omitempty" bson:"customDomain,omitempty" binding:"omitempty,domain_name"`

	// Zone is the zone that the deployment will be created in.
	// If the zone is not set, the deployment will be created in the default zone.
	Zone *string `json:"zone" bson:"zone,omitempty" binding:"omitempty"`
}

type DeploymentUpdate struct {
	Name *string `json:"name,omitempty" bson:"name,omitempty" binding:"omitempty,required,rfc1035,min=3,max=30,deployment_name"`

	CpuCores *float64         `json:"cpuCores,omitempty" bson:"cpuCores,omitempty" binding:"omitempty,min=0.1"`
	RAM      *float64         `json:"ram,omitempty" bson:"ram,omitempty" binding:"omitempty,min=0.1"`
	Replicas *int             `json:"replicas,omitempty" bson:"replicas,omitempty" binding:"omitempty,min=0,max=100"`
	GPUs     *[]DeploymentGPU `json:"gpus,omitempty" bson:"gpus,omitempty" binding:"omitempty,min=0"`

	Envs         *[]Env    `json:"envs,omitempty" bson:"envs,omitempty" binding:"omitempty,env_list,min=0,max=1000,dive"`
	Volumes      *[]Volume `json:"volumes,omitempty" bson:"volumes,omitempty" binding:"omitempty,min=0,max=100,dive"`
	InitCommands *[]string `json:"initCommands,omitempty" bson:"initCommands,omitempty" binding:"omitempty,min=0,max=100,dive,min=0,max=100"`
	Args         *[]string `json:"args,omitempty" bson:"args,omitempty" binding:"omitempty,min=0,max=100,dive,min=0,max=100"`
	Visibility   *string   `json:"visibility" bson:"visibility" binding:"omitempty,oneof=public private auth"`

	NeverStale *bool `json:"neverStale,omitempty" bson:"neverStale" binding:"omitempty,boolean"`

	// Deprecated: Use Visibility instead.
	Private *bool `json:"private,omitempty" bson:"private,omitempty" binding:"omitempty,boolean"`

	Image           *string `json:"image,omitempty" bson:"image,omitempty" binding:"omitempty,min=1,max=1000"`
	HealthCheckPath *string `json:"healthCheckPath,omitempty" bson:"healthCheckPath,omitempty" binding:"omitempty,min=0,max=1000,health_check_path"`
	// CustomDomain is the domain that the deployment will be available on.
	// The max length is set to 243 to allow for a subdomain when confirming the domain.
	CustomDomain *string `json:"customDomain,omitempty" bson:"customDomain,omitempty" binding:"omitempty,domain_name"`
}

type Env struct {
	Name  string `json:"name" bson:"name" binding:"required,env_name,min=1,max=100"`
	Value string `json:"value" bson:"value" binding:"required,min=1,max=10000"`
}

type Volume struct {
	Name       string `json:"name" bson:"name" binding:"required,volume_name,min=3,max=30"`
	AppPath    string `json:"appPath" bson:"appPath" binding:"required,min=1,max=255"`
	ServerPath string `json:"serverPath" bson:"serverPath" binding:"required,min=1,max=255"`
}

type DeploymentBuild struct {
	Name      string `bson:"name"`
	Tag       string `bson:"tag"`
	Branch    string `bson:"branch"`
	ImportURL string `bson:"importUrl"`
}

type ReplicaStatus struct {
	// DesiredReplicas is the number of replicas that the deployment should have.
	DesiredReplicas int `json:"desiredReplicas"`
	// ReadyReplicas is the number of replicas that are ready.
	ReadyReplicas int `json:"readyReplicas"`
	// AvailableReplicas is the number of replicas that are available.
	AvailableReplicas int `json:"availableReplicas"`
	// UnavailableReplicas is the number of replicas that are unavailable.
	UnavailableReplicas int `json:"unavailableReplicas"`
}

type DeploymentCreated struct {
	ID    string `json:"id"`
	JobID string `json:"jobId"`
}

type DeploymentDeleted struct {
	ID    string `json:"id"`
	JobID string `json:"jobId"`
}

type DeploymentUpdated struct {
	ID    string  `json:"id"`
	JobID *string `json:"jobId,omitempty"`
}

type DeploymentGPU struct {
	Name         string  `json:"name"`
	TemplateName *string `json:"templateName,omitempty"`
	ClaimName    *string `json:"claimName,omitempty"`
}

type DeploymentSpecs struct {
	CpuCores float64         `json:"cpuCores"`
	RAM      float64         `json:"ram"`
	Replicas int             `json:"replicas"`
	GPUs     []DeploymentGPU `json:"gpus,omitempty"`
}

type CiConfig struct {
	Config string `json:"config"`
}

type DeploymentCommand struct {
	Command string `json:"command" bson:"command" binding:"required,oneof=restart"`
}

type LogMessage struct {
	Source    string    `json:"source"`
	Prefix    string    `json:"prefix"`
	Line      string    `json:"line"`
	CreatedAt time.Time `json:"createdAt"`
}
