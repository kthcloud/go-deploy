package body

import "time"

type VmRead struct {
	ID      string  `json:"id"`
	Name    string  `json:"name"`
	OwnerID string  `json:"ownerId"`
	Zone    string  `json:"zone"`
	Host    *string `json:"host,omitempty"`

	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt,omitempty"`
	RepairedAt *time.Time `json:"repairedAt,omitempty"`

	Specs        Specs       `json:"specs,omitempty"`
	Ports        []PortRead  `json:"ports"`
	GPU          *VmGpuLease `json:"gpu_repo,omitempty"`
	SshPublicKey string      `json:"sshPublicKey"`

	Teams []string `json:"teams"`

	Status              string  `json:"status"`
	SshConnectionString *string `json:"sshConnectionString,omitempty"`
}

type VmCreate struct {
	Name         string       `json:"name" bson:"name" binding:"required,rfc1035,min=3,max=30"`
	SshPublicKey string       `json:"sshPublicKey" bson:"sshPublicKey" binding:"required,ssh_public_key"`
	Ports        []PortCreate `json:"ports" bson:"ports" binding:"omitempty,port_list_names,port_list_numbers,port_list_http_proxies,min=0,max=10,dive"`

	CpuCores int `json:"cpuCores" bson:"cpuCores" binding:"required,min=2"`
	RAM      int `json:"ram" bson:"ram" binding:"required,min=1"`
	DiskSize int `json:"diskSize" bson:"diskSize" binding:"required,min=20"`

	Zone *string `json:"zone,omitempty" bson:"zone,omitempty" binding:"omitempty"`
}

type VmUpdate struct {
	Ports    *[]PortUpdate `json:"ports,omitempty" bson:"ports,omitempty" binding:"omitempty,port_list_names,port_list_numbers,port_list_http_proxies,min=0,max=10,dive"`
	CpuCores *int          `json:"cpuCores,omitempty" bson:"cpuCores,omitempty" binding:"omitempty,min=1"`
	RAM      *int          `json:"ram,omitempty" bson:"ram,omitempty" binding:"omitempty,min=1"`

	// Name is used to rename a VM.
	// If specified, only name will be updated.
	Name *string `json:"name,omitempty" bson:"name,omitempty" binding:"omitempty,rfc1035,min=3,max=30"`

	// OwnerID is used to initiate transfer a VM to another user.
	// If specified, only the transfer will happen.
	// If specified but empty, the transfer will be canceled.
	OwnerID *string `json:"ownerId,omitempty" bson:"ownerId,omitempty" binding:"omitempty"`

	// TransferCode is used to accept transfer of a VM.
	// If specified, only the transfer will happen.
	TransferCode *string `json:"transferCode,omitempty" bson:"transferCode,omitempty" binding:"omitempty,min=1,max=1000"`

	// SnapshotID is used to apply snapshot to a VM.
	// If specified, only the snapshot application will happen.
	SnapshotID *string `json:"snapshotId,omitempty" bson:"snapshotId,omitempty" binding:"omitempty,uuid4"`

	// GpuID is used to attach/detach a GPU to a VM.
	// If specified and not empty, only the GPU will be attached.
	// If specified and empty, only the GPU will be detached.
	GpuID *string `json:"gpuId,omitempty" bson:"gpuId,omitempty" binding:"omitempty,min=0,max=100"`
}

type VmUpdateOwner struct {
	NewOwnerID   string  `json:"newOwnerId" bson:"newOwnerId" binding:"required,uuid4"`
	OldOwnerID   string  `json:"oldOwnerId" bson:"oldOwnerId" binding:"required,uuid4"`
	TransferCode *string `json:"transferCode,omitempty" bson:"transferCode,omitempty" binding:"omitempty,min=1,max=1000"`
}

type VmGpuLease struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	LeaseEndAt time.Time `json:"leaseEndAt"`
	IsExpired  bool      `json:"isExpired"`
}

type Specs struct {
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
