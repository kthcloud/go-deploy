package body

type PublicKey struct {
	Name string `json:"name" binding:"required,rfc1035,min=1,max=30"`
	Key  string `json:"key" binding:"required,ssh_public_key"`
}

type Quota struct {
	Deployment int `json:"deployment"`
	CpuCores   int `json:"cpuCores"`
	RAM        int `json:"ram"`
	DiskSpace  int `json:"diskSpace"`
}

type UserRead struct {
	ID         string      `json:"id"`
	Username   string      `json:"username"`
	Email      string      `json:"email"`
	Roles      []string    `json:"roles"`
	Quota      Quota       `json:"quota"`
	PublicKeys []PublicKey `json:"publicKeys"`
}

type UserUpdate struct {
	Username   *string      `json:"username" binding:"omitempty,min=3,max=32"`
	Email      *string      `json:"email" binding:"omitempty,email"`
	PublicKeys *[]PublicKey `json:"publicKeys" binding:"omitempty,dive,min=0,max=1000"`
}
