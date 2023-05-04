package body

type DeploymentPort struct {
	Name     string `json:"name" binding:"required,rfc1035"`
	Port     int    `json:"port" binding:"required,min=1,max=65535"`
	Protocol string `json:"protocol" binding:"required,oneof=tcp udp"`
}

type DeploymentCreate struct {
	Name    string              `json:"name" binding:"required,rfc1035,min=3,max=30"`
	Private bool                `json:"private" binding:"omitempty,boolean"`
	Envs    []map[string]string `json:"envs" binding:"omitempty,max=1000"`
}

type DeploymentUpdate struct {
	Envs *[]map[string]string `json:"envs" binding:"omitempty,max=1000"`
}

type DeploymentCreated struct {
	ID    string `json:"id"`
	JobID string `json:"jobId"`
}

type DeploymentDeleted struct {
	ID    string `json:"id"`
	JobID string `json:"jobId"`
}

type DeploymentRead struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	OwnerID string  `json:"ownerId"`
	Status  string  `json:"status"`
	URL     *string `json:"url,omitempty"`
}

type CIConfig struct {
	Config string `json:"config"`
}
