package body

type UserUpdate struct {
	Username        *string   `json:"username" binding:"omitempty,min=3,max=32"`
	Email           *string   `json:"email" binding:"omitempty,email"`
	VmQuota         *int      `json:"vmQuota" binding:"omitempty,min=0,max=1000"`
	DeploymentQuota *int      `json:"deploymentQuota" binding:"omitempty,min=0,max=1000"`
	PublicKeys      *[]string `json:"publicKeys" binding:"omitempty,min=0,max=1000"`
}
