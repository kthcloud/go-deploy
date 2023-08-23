package query

type Env struct {
	Key string `json:"key" binding:"required,env_name,min=1,max=100"`
	Val string `json:"val" binding:"required,min=1,max=10000"`
}

type DeploymentList struct {
	WantAll bool `form:"all" binding:"omitempty,boolean"`
}

type DeploymentUpdate struct {
	Envs []map[string]string `json:"envs" binding:"omitempty,dive,min=0,max=1000"`
}

type StorageManagerList struct {
	WantAll bool `form:"all" binding:"omitempty,boolean"`
}
