package body

type UserRead struct {
	ID              string   `json:"id"`
	Username        string   `json:"username"`
	Email           string   `json:"email"`
	Admin           bool     `json:"admin"`
	VmQuota         int      `json:"vmQuota"`
	DeploymentQuota int      `json:"deploymentQuota"`
	PowerUser       bool     `json:"powerUser"`
	PublicKeys      []string `json:"publicKeys"`
}

type UserUpdate struct {
	Username        *string   `json:"username" binding:"omitempty,min=3,max=32"`
	Email           *string   `json:"email" binding:"omitempty,email"`
	VmQuota         *int      `json:"vmQuota" binding:"omitempty,min=0,max=1000"`
	DeploymentQuota *int      `json:"deploymentQuota" binding:"omitempty,min=0,max=1000"`
	PublicKeys      *[]string `json:"publicKeys" binding:"omitempty,min=0,max=1000"`
}
