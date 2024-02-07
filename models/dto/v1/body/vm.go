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
	GPU          *VmGpuLease `json:"gpu,omitempty"`
	SshPublicKey string      `json:"sshPublicKey"`

	Teams []string `json:"teams"`

	Status           string  `json:"status"`
	ConnectionString *string `json:"connectionString,omitempty"`
}

type VmCreate struct {
	Name         string       `json:"name" bson:"name" binding:"required,rfc1035,min=3,max=30"`
	SshPublicKey string       `json:"sshPublicKey" bson:"sshPublicKey" binding:"required,ssh_public_key"`
	Ports        []PortCreate `json:"ports" bson:"ports" binding:"omitempty,port_list_names,port_list_numbers,port_list_http_proxies,min=0,max=10,dive"`
	CpuCores     int          `json:"cpuCores" bson:"cpuCores" binding:"required,min=2"`
	RAM          int          `json:"ram" bson:"ram" binding:"required,min=1"`
	DiskSize     int          `json:"diskSize" bson:"diskSize" binding:"required,min=20"`
	Zone         *string      `json:"zone,omitempty" bson:"zone,omitempty" binding:"omitempty"`
}

type VmUpdate struct {
	Name       *string       `json:"name,omitempty" bson:"name,omitempty" binding:"omitempty,rfc1035,min=3,max=30"`
	SnapshotID *string       `json:"snapshotId,omitempty" bson:"snapshotId,omitempty" binding:"omitempty,uuid4"`
	Ports      *[]PortUpdate `json:"ports,omitempty" bson:"ports,omitempty" binding:"omitempty,port_list_names,port_list_numbers,port_list_http_proxies,min=0,max=10,dive"`
	CpuCores   *int          `json:"cpuCores,omitempty" bson:"cpuCores,omitempty" binding:"omitempty,min=1"`
	RAM        *int          `json:"ram,omitempty" bson:"ram,omitempty" binding:"omitempty,min=1"`

	GpuID      *string `json:"gpuId,omitempty" bson:"gpuId,omitempty" binding:"omitempty,min=0,max=100"`
	NoLeaseEnd *bool   `json:"noLeaseEnd,omitempty" bson:"noLeaseEnd,omitempty" binding:"omitempty"`

	OwnerID      *string `json:"ownerId,omitempty" bson:"ownerId,omitempty" binding:"omitempty"`
	TransferCode *string `json:"transferCode,omitempty" bson:"transferCode,omitempty" binding:"omitempty,min=1,max=1000"`
}

type VmUpdateOwner struct {
	NewOwnerID   string  `json:"newOwnerId" bson:"newOwnerId" binding:"required,uuid4"`
	OldOwnerID   string  `json:"oldOwnerId" bson:"oldOwnerId" binding:"required,uuid4"`
	TransferCode *string `json:"transferCode,omitempty" bson:"transferCode,omitempty" binding:"omitempty,min=1,max=1000"`
}

type VmGpuLease struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	LeaseEnd     time.Time `json:"leaseEnd"`
	LeaseExpired bool      `json:"expired"`
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
