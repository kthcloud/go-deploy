package body

type PublicKey struct {
	Name string `json:"name" binding:"required,min=1,max=30"`
	Key  string `json:"key" binding:"required,ssh_public_key"`
}

type Quota struct {
	Deployments int `json:"deployments"`
	CpuCores    int `json:"cpuCores"`
	RAM         int `json:"ram"`
	DiskSize    int `json:"diskSize"`
	Snapshots   int `json:"snapshots"`
}

type Role struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

type UserRead struct {
	ID         string      `json:"id"`
	Username   string      `json:"username"`
	Email      string      `json:"email"`
	PublicKeys []PublicKey `json:"publicKeys"`

	Role  Role `json:"role"`
	Admin bool `json:"admin"`

	Quota Quota `json:"quota"`
	Usage Quota `json:"usage"`

	StorageURL *string `json:"storageUrl,omitempty"`
}

type UserUpdate struct {
	Username   *string      `json:"username" binding:"omitempty,min=3,max=32"`
	PublicKeys *[]PublicKey `json:"publicKeys" binding:"omitempty,dive,min=0,max=1000"`
}
