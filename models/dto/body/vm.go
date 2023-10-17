package body

import "time"

type VmHttpProxy struct {
	Name         string  `json:"name" binding:"required,rfc1035,min=3,max=30"`
	CustomDomain *string `json:"customDomain,omitempty" binding:"omitempty,domain_name,min=1,max=253"`

	URL             *string `json:"url,omitempty" `
	CustomDomainURL *string `json:"customDomainUrl,omitempty" `
}

type PortRead struct {
	Name         string       `json:"name,omitempty" binding:"required,min=1,max=100"`
	Port         int          `json:"port,omitempty"  binding:"required,min=1,max=65535"`
	ExternalPort *int         `json:"externalPort,omitempty"`
	Protocol     string       `json:"protocol,omitempty" binding:"required,oneof=tcp udp"`
	HttpProxy    *VmHttpProxy `json:"httpProxy,omitempty" binding:"omitempty,dive"`
}

type Specs struct {
	CpuCores int `json:"cpuCores,omitempty"`
	RAM      int `json:"ram,omitempty"`
	DiskSize int `json:"diskSize,omitempty"`
}

type VmGpu struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	LeaseEnd     time.Time `json:"leaseEnd"`
	LeaseExpired bool      `json:"expired"`
}

type VmRead struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	OwnerID string `json:"ownerId"`
	Zone    string `json:"zone"`

	Specs        Specs      `json:"specs,omitempty"`
	Ports        []PortRead `json:"ports"`
	GPU          *VmGpu     `json:"gpu,omitempty"`
	SshPublicKey string     `json:"sshPublicKey"`

	Status           string  `json:"status"`
	ConnectionString *string `json:"connectionString,omitempty"`
}

type VmCreate struct {
	Name         string     `json:"name" binding:"required,rfc1035,min=3,max=30"`
	SshPublicKey string     `json:"sshPublicKey" binding:"required,ssh_public_key"`
	Ports        []PortRead `json:"ports" binding:"omitempty,port_list_names,port_list_numbers,port_list_http_proxies,min=0,max=100,dive"`
	CpuCores     int        `json:"cpuCores" binding:"required,min=2"`
	RAM          int        `json:"ram" binding:"required,min=1"`
	DiskSize     int        `json:"diskSize" binding:"required,min=20"`
	Zone         *string    `json:"zone" binding:"omitempty"`
}

type VmUpdate struct {
	SnapshotID *string     `json:"snapshotId" binding:"omitempty,uuid4"`
	GpuID      *string     `json:"gpuId" binding:"omitempty,min=0,max=100"`
	Ports      *[]PortRead `json:"ports" binding:"omitempty,port_list_names,port_list_numbers,port_list_http_proxies,min=0,max=1000,dive"`
	CpuCores   *int        `json:"cpuCores" binding:"omitempty,min=1"`
	RAM        *int        `json:"ram" binding:"omitempty,min=1"`
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

type VmSnapshotCreated struct {
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
	VmID    *string   `json:"vmId,omitempty"`
	User    *string   `json:"user,omitempty"`
	End     time.Time `json:"end"`
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
