package vm

import (
	"go-deploy/models/sys/vm/subsystems"
	"time"
)

const (
	ActivityBeingCreated     = "beingCreated"
	ActivityBeingDeleted     = "beingDeleted"
	ActivityBeingUpdated     = "beingUpdated"
	ActivityAttachingGPU     = "attachingGpu"
	ActivityDetachingGPU     = "detachingGpu"
	ActivityRepairing        = "repairing"
	ActivityCreatingSnapshot = "creatingSnapshot"
	ActivityApplyingSnapshot = "applyingSnapshot"
)

type Port struct {
	Name     string `bson:"name"`
	Port     int    `bson:"port"`
	Protocol string `bson:"protocol"`
}

type Subsystems struct {
	CS subsystems.CS `bson:"cs"`
}

type Usage struct {
	CpuCores  int `json:"cpuCores"`
	RAM       int `json:"ram"`
	DiskSize  int `json:"diskSize"`
	Snapshots int `json:"snapshots"`
}

type CreateParams struct {
	Name string `json:"name"`
	Zone string `json:"zone"`

	NetworkID *string `json:"networkId"`

	SshPublicKey string `json:"sshPublicKey"`
	Ports        []Port `json:"ports"`

	CpuCores int `json:"cpuCores"`
	RAM      int `json:"ram"`
	DiskSize int `json:"diskSize"`
}

type UpdateParams struct {
	SnapshotID *string `json:"snapshotId"`
	Ports      *[]Port `json:"ports"`
	CpuCores   *int    `json:"cpuCores"`
	RAM        *int    `json:"ram"`
}

type Snapshot struct {
	ID         string    `json:"id"`
	VmID       string    `json:"vmId"`
	Name       string    `json:"displayname"`
	ParentName *string   `json:"parentName,omitempty"`
	CreatedAt  time.Time `json:"created"`
	State      string    `json:"state"`
	Current    bool      `json:"current"`
}
