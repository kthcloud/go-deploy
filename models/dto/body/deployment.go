package body

import "time"

type Env struct {
	Name  string `json:"name" bson:"name" binding:"required,env_name,min=1,max=100"`
	Value string `json:"value" bson:"value" binding:"required,min=1,max=10000"`
}

type Volume struct {
	Name       string `json:"name" bson:"name" binding:"required,min=1,max=100"`
	AppPath    string `json:"appPath" bson:"appPath" binding:"required,min=1,max=100"`
	ServerPath string `json:"serverPath" bson:"serverPath" binding:"required,min=1,max=100"`
}

type GitHub struct {
	Token        string `json:"token" bson:"token" binding:"required,min=1,max=1000"`
	RepositoryID int64  `json:"repositoryId" bson:"repositoryId" binding:"required"`
}

type DeploymentCreate struct {
	Name string `json:"name" bson:"name" binding:"required,rfc1035,min=3,max=30"`

	Image           *string  `json:"image,omitempty" bson:"image,omitempty" binding:"omitempty,min=1,max=1000"`
	Private         bool     `json:"private" bson:"private" binding:"omitempty,boolean"`
	Envs            []Env    `json:"envs" bson:"envs" binding:"omitempty,env_list,min=0,max=1000,dive"`
	Volumes         []Volume `json:"volumes" bson:"volumes" binding:"omitempty,min=0,max=100,dive"`
	InitCommands    []string `json:"initCommands" bson:"initCommands" binding:"omitempty,min=0,max=100,dive,min=0,max=100"`
	HealthCheckPath *string  `json:"healthCheckPath" bson:"healthCheckPath,omitempty" binding:"omitempty,min=0,max=1000,health_check_path"`
	CustomDomain    *string  `json:"customDomain" bson:"customDomain,omitempty" binding:"omitempty,domain_name,custom_domain,min=1,max=253"`
	Replicas        *int     `json:"replicas" bson:"replicas,omitempty" binding:"omitempty,min=0,max=100"`

	GitHub *GitHub `json:"github" bson:"github,omitempty" binding:"omitempty,dive"`

	Zone *string `json:"zone" bson:"zone,omitempty" binding:"omitempty"`
}

type DeploymentUpdate struct {
	// update
	Name            *string   `json:"name,omitempty" bson:"name,omitempty" binding:"omitempty,required,rfc1035,min=3,max=30"`
	Private         *bool     `json:"private,omitempty" bson:"private,omitempty" binding:"omitempty,boolean"`
	Envs            *[]Env    `json:"envs,omitempty" bson:"envs,omitempty" binding:"omitempty,env_list,min=0,max=1000,dive"`
	Volumes         *[]Volume `json:"volumes,omitempty" bson:"volumes,omitempty" binding:"omitempty,min=0,max=100,dive"`
	InitCommands    *[]string `json:"initCommands,omitempty" bson:"initCommands,omitempty" binding:"omitempty,min=0,max=100,dive"`
	CustomDomain    *string   `json:"customDomain,omitempty" bson:"customDomain,omitempty" binding:"omitempty,domain_name,custom_domain,min=0,max=253"`
	Image           *string   `json:"image,omitempty,omitempty" bson:"image,omitempty" binding:"omitempty,min=1,max=1000"`
	HealthCheckPath *string   `json:"healthCheckPath,omitempty" bson:"healthCheckPath,omitempty" binding:"omitempty,min=0,max=1000,health_check_path"`
	Replicas        *int      `json:"replicas,omitempty" bson:"replicas,omitempty" binding:"omitempty,min=0,max=100"`

	// update owner
	OwnerID      *string `json:"ownerId,omitempty" bson:"ownerId,omitempty" binding:"omitempty"`
	TransferCode *string `json:"transferCode,omitempty" bson:"transferCode,omitempty" binding:"omitempty,min=1,max=1000"`
}

type DeploymentUpdateOwner struct {
	NewOwnerID   string  `json:"newOwnerId" bson:"newOwnerId" binding:"required,uuid4"`
	OldOwnerID   string  `json:"oldOwnerId" bson:"oldOwnerId" binding:"required,uuid4"`
	TransferCode *string `json:"transferCode,omitempty" bson:"transferCode,omitempty" binding:"omitempty,min=1,max=1000"`
}

type DeploymentBuild struct {
	Name      string `bson:"name"`
	Tag       string `bson:"tag"`
	Branch    string `bson:"branch"`
	ImportURL string `bson:"importUrl"`
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

type DeploymentRead struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Type        string     `json:"type"`
	OwnerID     string     `json:"ownerId"`
	Zone        string     `json:"zone"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty"`
	RepairedAt  *time.Time `json:"repairedAt,omitempty"`
	RestartedAt *time.Time `json:"restartedAt,omitempty"`

	URL             *string  `json:"url,omitempty"`
	CustomDomainURL *string  `json:"customDomainUrl,omitempty"`
	Envs            []Env    `json:"envs"`
	Volumes         []Volume `json:"volumes"`
	InitCommands    []string `json:"initCommands"`
	Private         bool     `json:"private"`
	InternalPort    int      `json:"internalPort"`
	Image           *string  `json:"image,omitempty"`
	HealthCheckPath *string  `json:"healthCheckPath,omitempty"`
	Replicas        int      `json:"replicas"`

	Status     string `json:"status"`
	PingResult *int   `json:"pingResult,omitempty"`

	Integrations []string `json:"integrations"`

	StorageURL *string `json:"storageUrl,omitempty"`
}

type CiConfig struct {
	Config string `json:"config"`
}

type DeploymentCommand struct {
	Command string `json:"command" bson:"command" binding:"required,oneof=restart"`
}
