package body

import "time"

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
	Onboarded  bool        `json:"onboarded"`

	Role  Role `json:"role"`
	Admin bool `json:"admin"`

	Quota Quota `json:"quota"`
	Usage Quota `json:"usage"`

	StorageURL *string `json:"storageUrl,omitempty"`
}

type UserUpdate struct {
	Username   *string      `json:"username,omitempty" bson:"username,omitempty" binding:"omitempty,min=3,max=32"`
	PublicKeys *[]PublicKey `json:"publicKeys,omitempty" bson:"publicKeys,omitempty" binding:"omitempty,min=0,max=1000,dive"`
	Onboarded  *bool        `json:"onboarded,omitempty" bson:"onboarded,omitempty" binding:"omitempty,boolean"`
}

type TeamMember struct {
	UserRead
	TeamRole string    `json:"teamRole"`
	JoinedAt time.Time `json:"joinedAt"`
}

type TeamMemberCreate struct {
	ID       string `json:"id" binding:"required,uuid4"`
	TeamRole string `json:"teamRole"`
}

type TeamMemberUpdate struct {
	ID       string `json:"id" binding:"required,uuid4"`
	TeamRole string `json:"teamRole"`
}

type TeamCreate struct {
	Name        string             `json:"name" binding:"required,min=3,max=32"`
	Description string             `json:"description" binding:"required,min=3,max=1000"`
	Members     []TeamMemberCreate `json:"members" binding:"required,min=1,max=1000"`
}

type TeamUpdate struct {
	Name        *string             `json:"name,omitempty" binding:"omitempty,min=3,max=32"`
	Description *string             `json:"description,omitempty" binding:"omitempty,min=3,max=1000"`
	Members     *[]TeamMemberUpdate `json:"members,omitempty" binding:"omitempty,min=1,max=1000"`
}

type TeamRead struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Members     []TeamMember `json:"members"`
}
