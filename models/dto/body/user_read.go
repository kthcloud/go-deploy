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
