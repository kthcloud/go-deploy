package body

import "time"

type VmPort struct {
	Name     string `json:"name,omitempty" bson:"name" binding:"required,rfc1035"`
	Port     int    `json:"port,omitempty" bson:"port" binding:"required,min=1,max=65535"`
	Protocol string `json:"protocol,omitempty" bson:"protocol" binding:"required,oneof=tcp udp"`
}

type VmGpu struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	LeaseEnd time.Time `json:"leaseEnd"`
}

type VmRead struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	SshPublicKey     string `json:"sshPublicKey"`
	OwnerID          string `json:"ownerId"`
	Status           string `json:"status"`
	ConnectionString string `json:"connectionString"`
	GPU              *VmGpu `json:"gpu,omitempty"`
}

type VmCreate struct {
	Name         string `json:"name" binding:"required,rfc1035,min=3,max=30"`
	SshPublicKey string `json:"sshPublicKey" binding:"required,ssh_public_key"`
}

type VmUpdate struct {
	Ports *[]VmPort `json:"ports" bson:"ports" binding:"omitempty,dive"`
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
	VmID string    `bson:"vmId" json:"vmId"`
	User string    `bson:"user" json:"user"`
	End  time.Time `bson:"end" json:"end"`
}

type GpuRead struct {
	ID    string    `json:"id"`
	Name  string    `json:"name"`
	Lease *GpuLease `json:"lease,omitempty"`
}

type DoCommand struct {
	Command string `json:"command" binding:"required,oneof=start stop reboot"`
}
