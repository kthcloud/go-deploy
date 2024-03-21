package query

type Env struct {
	Key string `json:"key" binding:"required,env_name,min=1,max=100"`
	Val string `json:"val" binding:"required,min=1,max=10000"`
}

type DeploymentList struct {
	*Pagination

	All    bool    `form:"all" binding:"omitempty,boolean"`
	UserID *string `form:"userId" binding:"omitempty,uuid4"`
}

type DeploymentUpdate struct {
	Envs []map[string]string `json:"envs" binding:"omitempty,dive,min=0,max=1000"`
}
