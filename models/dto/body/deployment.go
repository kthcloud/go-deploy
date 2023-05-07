package body

type Env struct {
	Name  string `json:"name" binding:"required,env_name,min=1,max=100"`
	Value string `json:"value" binding:"required,min=1,max=10000"`
}

type DeploymentCreate struct {
	Name    string `json:"name" binding:"required,rfc1035,min=3,max=30"`
	Private bool   `json:"private" binding:"omitempty,boolean"`
	Envs    []Env  `json:"envs" binding:"omitempty,env_list,dive,min=0,max=1000"`
}

type DeploymentUpdate struct {
	Private *bool  `json:"private" binding:"omitempty,boolean"`
	Envs    *[]Env `json:"envs" binding:"omitempty,env_list,dive,min=0,max=1000"`
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
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	OwnerID string  `json:"ownerId"`
	Status  string  `json:"status"`
	URL     *string `json:"url,omitempty"`
	Envs    []Env   `json:"envs"`
	Private bool    `json:"private"`
}

type CiConfig struct {
	Config string `json:"config"`
}
