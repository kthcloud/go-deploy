package body

import "time"

type VmRead struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	InternalName *string `json:"internalName,omitempty"`
	OwnerID      string  `json:"ownerId"`
	Zone         string  `json:"zone"`
	Host         *string `json:"host,omitempty"`

	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt,omitempty"`
	RepairedAt *time.Time `json:"repairedAt,omitempty"`
	AccessedAt time.Time  `json:"accessedAt"`

	NeverStale bool `json:"neverStale"`

	Specs        VmSpecs     `json:"specs"`
	Ports        []PortRead  `json:"ports"`
	GPU          *VmGpuLease `json:"gpu,omitempty"`
	SshPublicKey string      `json:"sshPublicKey"`

	Teams []string `json:"teams"`

	Status              string  `json:"status"`
	SshConnectionString *string `json:"sshConnectionString,omitempty"`
}

type VmCreate struct {
	Name         string       `json:"name" bson:"name" binding:"required,rfc1035,min=3,max=30,vm_name"`
	SshPublicKey string       `json:"sshPublicKey" bson:"sshPublicKey" binding:"required,ssh_public_key"`
	Ports        []PortCreate `json:"ports" bson:"ports" binding:"omitempty,port_list_names,port_list_numbers,port_list_http_proxies,min=0,max=10,dive"`

	CpuCores int `json:"cpuCores" bson:"cpuCores" binding:"required,min=1"`
	RAM      int `json:"ram" bson:"ram" binding:"required,min=1"`
	DiskSize int `json:"diskSize" bson:"diskSize" binding:"required,min=10"`

	Zone *string `json:"zone,omitempty" bson:"zone,omitempty" binding:"omitempty"`

	NeverStale bool `json:"neverStale" bson:"neverStale" binding:"omitempty,boolean"`
}

type VmUpdate struct {
	Name       *string       `json:"name,omitempty" bson:"name,omitempty" binding:"omitempty,rfc1035,min=3,max=30,vm_name"`
	Ports      *[]PortUpdate `json:"ports,omitempty" bson:"ports,omitempty" binding:"omitempty,port_list_names,port_list_numbers,port_list_http_proxies,min=0,max=10,dive"`
	CpuCores   *int          `json:"cpuCores,omitempty" bson:"cpuCores,omitempty" binding:"omitempty,min=1"`
	RAM        *int          `json:"ram,omitempty" bson:"ram,omitempty" binding:"omitempty,min=1"`
	NeverStale *bool         `json:"neverStale,omitempty" bson:"neverStale" binding:"omitempty,boolean"`
}

type VmUpdateOwner struct {
	NewOwnerID string `json:"newOwnerId" bson:"newOwnerId" binding:"required,uuid4"`
	OldOwnerID string `json:"oldOwnerId" bson:"oldOwnerId" binding:"required,uuid4"`
}

type VmGpuLease struct {
	ID            string  `json:"id"`
	GpuGroupID    string  `json:"gpuGroupId"`
	LeaseDuration float64 `json:"leaseDuration"`
	// ActivatedAt specifies the time when the lease was activated. This is the time the user first attached the GPU
	// or 1 day after the lease was created if the user did not attach the GPU.
	ActivatedAt *time.Time `json:"activatedAt,omitempty"`
	// AssignedAt specifies the time when the lease was assigned to the user.
	AssignedAt *time.Time `json:"assignedAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	// ExpiresAt specifies the time when the lease will expire.
	// This is only present if the lease is active.
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	// ExpiredAt specifies the time when the lease expired.
	// This is only present if the lease is expired.
	ExpiredAt *time.Time `json:"expiredAt,omitempty"`
}

type VmSpecs struct {
	CpuCores int `json:"cpuCores,omitempty"`
	RAM      int `json:"ram,omitempty"`
	DiskSize int `json:"diskSize,omitempty"`
}

type VmCreated struct {
	ID    string `json:"id"`
	JobID string `json:"jobId"`
}

type VmDeleted struct {
	ID    string `json:"id"`
	JobID string `json:"jobId"`
}

type VmUpdated struct {
	ID    string  `json:"id"`
	JobID *string `json:"jobId,omitempty"`
}
