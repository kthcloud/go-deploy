package body

import "time"

type Env struct {
	Name  string `json:"name" binding:"required,env_name,min=1,max=100"`
	Value string `json:"value" binding:"required,min=1,max=10000"`
}

type Volume struct {
	Name       string `json:"name" binding:"required,min=1,max=100"`
	AppPath    string `json:"appPath" binding:"required,min=1,max=100"`
	ServerPath string `json:"serverPath" binding:"required,min=1,max=100"`
}

type DeploymentCreate struct {
	Name string `json:"name" binding:"required,rfc1035,min=3,max=30"`

	Private      bool     `json:"private" binding:"omitempty,boolean"`
	Envs         []Env    `json:"envs" binding:"omitempty,env_list,dive,min=0,max=1000"`
	Volumes      []Volume `json:"volumes" binding:"omitempty,dive,min=0,max=100"`
	InitCommands []string `json:"initCommands" binding:"omitempty,dive,min=0,max=100"`

	GitHub *struct {
		Token        string `json:"token" binding:"required,min=1,max=1000"`
		RepositoryID int64  `json:"repositoryId" binding:"required"`
	} `json:"github" binding:"omitempty,dive"`

	Zone *string `json:"zone" binding:"omitempty"`
}

type DeploymentUpdate struct {
	Private      *bool     `json:"private" binding:"omitempty,boolean"`
	Envs         *[]Env    `json:"envs" binding:"omitempty,env_list,dive,min=0,max=1000"`
	Volumes      *[]Volume `json:"volumes" binding:"omitempty,dive,min=0,max=100"`
	InitCommands *[]string `json:"initCommands" binding:"omitempty,dive,min=0,max=100"`
	ExtraDomains *[]string `json:"extraDomains" binding:"omitempty,extra_domain_list,dive,min=0,max=1000"`
}

type DeploymentBuild struct {
	Tag       string `json:"tag" binding:"required,rfc1035,min=1,max=50"`
	Branch    string `json:"branch" binding:"required,min=1,max=50"`
	ImportURL string `json:"importUrl" binding:"required,min=1,max=1000"`
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
	ID    string `json:"id"`
	JobID string `json:"jobId"`
}

type DeploymentRead struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	OwnerID string `json:"ownerId"`
	Zone    string `json:"zone"`

	URL          *string  `json:"url,omitempty"`
	Envs         []Env    `json:"envs"`
	Volumes      []Volume `json:"volumes"`
	InitCommands []string `json:"initCommands"`
	Private      bool     `json:"private"`

	Status     string `json:"status"`
	PingResult *int   `json:"pingResult,omitempty"`

	Integrations []string `json:"integrations"`

	StorageURL *string `json:"storageUrl,omitempty"`
}

type CiConfig struct {
	Config string `json:"config"`
}

type DeploymentCommand struct {
	Command string `json:"command" binding:"required,oneof=restart"`
}

type StorageManagerDeleted struct {
	ID    string `json:"id"`
	JobID string `json:"jobId"`
}

type StorageManagerRead struct {
	ID        string    `json:"id"`
	OwnerID   string    `json:"ownerId"`
	CreatedAt time.Time `json:"createdAt"`
	Zone      string    `json:"zone"`
	URL       *string   `json:"url,omitempty"`
}
