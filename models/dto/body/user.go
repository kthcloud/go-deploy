package body

type PublicKey struct {
	Name string `json:"name" binding:"required,rfc1035,min=1,max=30"`
	Key  string `json:"key" binding:"required,ssh_public_key"`
}

type UserRead struct {
	ID              string      `json:"id"`
	Username        string      `json:"username"`
	Email           string      `json:"email"`
	Admin           bool        `json:"admin"`
	VmQuota         int         `json:"vmQuota"`
	DeploymentQuota int         `json:"deploymentQuota"`
	PowerUser       bool        `json:"powerUser"`
	PublicKeys      []PublicKey `json:"publicKeys"`
}

type UserUpdate struct {
	Username        *string      `json:"username" binding:"omitempty,min=3,max=32"`
	Email           *string      `json:"email" binding:"omitempty,email"`
	VmQuota         *int         `json:"vmQuota" binding:"omitempty,min=0,max=1000"`
	DeploymentQuota *int         `json:"deploymentQuota" binding:"omitempty,min=0,max=1000"`
	PublicKeys      *[]PublicKey `json:"publicKeys" binding:"omitempty,dive,min=0,max=1000"`
}
