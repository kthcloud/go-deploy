package body

import "time"

type Port struct {
	Name         string `json:"name,omitempty" bson:"name" binding:"required"`
	Port         int    `json:"port,omitempty" bson:"port" binding:"required,min=1,max=65535"`
	ExternalPort int    `json:"externalPort,omitempty" bson:"externalPort"`
	Protocol     string `json:"protocol,omitempty" bson:"protocol" binding:"required,oneof=tcp udp"`
}

type Specs struct {
	CpuCores int `json:"cpuCores,omitempty" bson:"cpuCores"`
	RAM      int `json:"ram,omitempty" bson:"ram"`
	DiskSize int `json:"diskSize,omitempty" bson:"diskSize"`
}

type VmGpu struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	LeaseEnd     time.Time `json:"leaseEnd"`
	LeaseExpired bool      `json:"expired"`
}

type VmRead struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	SshPublicKey     string  `json:"sshPublicKey"`
	Ports            []Port  `json:"ports"`
	OwnerID          string  `json:"ownerId"`
	Status           string  `json:"status"`
	ConnectionString *string `json:"connectionString,omitempty"`
	GPU              *VmGpu  `json:"gpu,omitempty"`
	Specs            Specs   `json:"specs,omitempty"`
}

type VmCreate struct {
	Name         string  `json:"name" binding:"required,rfc1035,min=3,max=30"`
	SshPublicKey string  `json:"sshPublicKey" binding:"required,ssh_public_key"`
	Ports        []Port  `json:"ports" binding:"omitempty,port_list_names,port_list_numbers,dive,min=0,max=1000"`
	CpuCores     int     `json:"cpuCores" binding:"required,min=1"`
	RAM          int     `json:"ram" binding:"required,min=1"`
	DiskSize     int     `json:"diskSize" binding:"required,min=20"`
	ZoneID       *string `json:"zoneId" binding:"omitempty,uuid4"`
}

type VmUpdate struct {
	SnapshotID *string `json:"snapshotId" binding:"omitempty,uuid4"`
	Ports      *[]Port `json:"ports" bson:"ports" binding:"omitempty,port_list_names,port_list_numbers,dive,min=0,max=1000"`
	CpuCores   *int    `json:"cpuCores" binding:"omitempty,min=1"`
	RAM        *int    `json:"ram" binding:"omitempty,min=1"`
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
	ID    string `json:"id"`
	JobID string `json:"jobId"`
}

type GpuAttached struct {
	ID    string `json:"id"`
	JobID string `json:"jobId"`
}

type GpuDetached struct {
	ID    string `json:"id"`
	JobID string `json:"jobId"`
}

type GpuLease struct {
	VmID    *string   `bson:"vmId" json:"vmId,omitempty"`
	User    *string   `bson:"user" json:"user,omitempty"`
	End     time.Time `bson:"end" json:"end"`
	Expired bool      `json:"expired"`
}

type GpuRead struct {
	ID    string    `json:"id"`
	Name  string    `json:"name"`
	Lease *GpuLease `json:"lease,omitempty"`
}

type VmCommand struct {
	Command string `json:"command" binding:"required,oneof=start stop reboot"`
}

type VmSnapshotRead struct {
	ID         string    `json:"id"`
	VmID       string    `json:"vmId"`
	Name       string    `json:"displayname"`
	ParentName *string   `json:"parentName,omitempty"`
	CreatedAt  time.Time `json:"created"`
	State      string    `json:"state"`
	Current    bool      `json:"current"`
}

type VmSnapshotCreate struct {
	Name string `json:"name" binding:"required,rfc1035,min=3,max=30"`
}
