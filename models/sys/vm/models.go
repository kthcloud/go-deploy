package vm

import (
	"go-deploy/models/sys/vm/subsystems"
	"time"
)

type PortHttpProxy struct {
	Name         string  `bson:"name"`
	CustomDomain *string `bson:"customDomain"`
}

type Port struct {
	Name      string         `bson:"name"`
	Port      int            `bson:"port"`
	Protocol  string         `bson:"protocol"`
	HttpProxy *PortHttpProxy `bson:"httpProxy"`
}

type Subsystems struct {
	CS  subsystems.CS  `bson:"cs"`
	K8s subsystems.K8s `bson:"k8s"`
}

type Usage struct {
	CpuCores  int `bson:"cpuCores"`
	RAM       int `bson:"ram"`
	DiskSize  int `bson:"diskSize"`
	Snapshots int `bson:"snapshots"`
}

type CreateParams struct {
	Name           string
	Zone           string
	DeploymentZone *string

	NetworkID *string

	SshPublicKey string
	Ports        []Port

	CpuCores int
	RAM      int
	DiskSize int
}

type UpdateParams struct {
	Name       *string
	OwnerID    *string
	SnapshotID *string
	Ports      *[]Port
	CpuCores   *int
	RAM        *int
}

type Snapshot struct {
	ID         string    `bson:"id"`
	VmID       string    `bson:"vmId"`
	Name       string    `bson:"displayname"`
	ParentName *string   `bson:"parentName,omitempty"`
	CreatedAt  time.Time `bson:"created"`
	State      string    `bson:"state"`
	Current    bool      `bson:"current"`
}

type CreateSnapshotParams struct {
	Name        string `bson:"name"`
	UserCreated bool   `bson:"userCreated"`
	Overwrite   bool   `bson:"overwrite"`
}
