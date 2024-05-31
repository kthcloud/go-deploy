package body

import "time"

type UserRead struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`

	PublicKeys []PublicKey `json:"publicKeys"`
	ApiKeys    []ApiKey    `json:"apiKeys"`
	UserData   []UserData  `json:"userData"`

	Role  Role `json:"role"`
	Admin bool `json:"admin"`

	Quota Quota `json:"quota"`
	Usage Usage `json:"usage"`

	StorageURL  *string `json:"storageUrl,omitempty"`
	GravatarURL *string `json:"gravatarUrl,omitempty"`
}

type UserReadDiscovery struct {
	ID          string  `json:"id"`
	Username    string  `json:"username"`
	FirstName   string  `json:"firstName"`
	LastName    string  `json:"lastName"`
	Email       string  `json:"email"`
	GravatarURL *string `json:"gravatarUrl,omitempty"`
}

type UserUpdate struct {
	PublicKeys *[]PublicKey `json:"publicKeys,omitempty" bson:"publicKeys,omitempty" binding:"omitempty,min=0,max=100,dive"`
	// ApiKeys specifies the API keys that should remain. If an API key is not in this list, it will be deleted.
	// However, API keys cannot be created, use /apiKeys endpoint to create new API keys.
	ApiKeys  *[]ApiKey   `json:"apiKeys,omitempty" bson:"apiKeys,omitempty" binding:"omitempty,min=0,max=100,dive"`
	UserData *[]UserData `json:"userData,omitempty" bson:"publicKeys,omitempty" binding:"omitempty,min=0,max=100,dive"`
}

type UserData struct {
	Key   string `json:"key" binding:"required,min=1,max=255"`
	Value string `json:"value" binding:"required,min=1,max=255"`
}

type PublicKey struct {
	Name string `json:"name" binding:"required,min=1,max=30"`
	Key  string `json:"key" binding:"required,ssh_public_key"`
}

type ApiKey struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	ExpiresAt time.Time `json:"expiresAt"`
}

type Quota struct {
	CpuCores         float64 `json:"cpuCores"`
	RAM              float64 `json:"ram"`
	DiskSize         float64 `json:"diskSize"`
	Snapshots        int     `json:"snapshots"`
	GpuLeaseDuration float64 `json:"gpuLeaseDuration"` // in hours
}

type Usage struct {
	CpuCores float64 `json:"cpuCores"`
	RAM      float64 `json:"ram"`
	DiskSize int     `json:"diskSize"`
}
