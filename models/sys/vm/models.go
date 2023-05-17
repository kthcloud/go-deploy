package vm

import (
	csModels "go-deploy/pkg/subsystems/cs/models"
)

const (
	ActivityBeingCreated = "beingCreated"
	ActivityBeingDeleted = "beingDeleted"
	ActivityBeingUpdated = "beingUpdated"
	ActivityAttachingGPU = "attachingGpu"
	ActivityDetachingGPU = "detachingGpu"
)

type Port struct {
	Name     string `bson:"name"`
	Port     int    `bson:"port"`
	Protocol string `bson:"protocol"`
}

type VM struct {
	ID        string `bson:"id"`
	Name      string `bson:"name"`
	OwnerID   string `bson:"ownerId"`
	ManagedBy string `bson:"managedBy"`

	GpuID        string   `bson:"gpuId"`
	SshPublicKey string   `bson:"sshPublicKey"`
	Ports        []Port   `bson:"ports"`
	Activities   []string `bson:"activities"`

	Subsystems    Subsystems `bson:"subsystems"`
	StatusCode    int        `bson:"statusCode"`
	StatusMessage string     `bson:"statusMessage"`
}

type Subsystems struct {
	CS CS `bson:"cs"`
}

type CS struct {
	ServiceOffering       csModels.ServiceOfferingPublic               `bson:"serviceOffering"`
	VM                    csModels.VmPublic                            `bson:"vm"`
	PortForwardingRuleMap map[string]csModels.PortForwardingRulePublic `bson:"portForwardingRuleMap"`
}

type Usage struct {
	CpuCores  int `json:"cpuCores"`
	RAM       int `json:"ram"`
	DiskSpace int `json:"diskSpace"`
}

type CreateParams struct {
	Name         string `json:"name"`
	SshPublicKey string `json:"sshPublicKey"`
	Ports        []Port `json:"ports"`
	CpuCores     int    `json:"cpuCores"`
	RAM          int    `json:"ram"`
	DiskSize     int    `json:"diskSize"`
}

type UpdateParams struct {
	Ports *[]Port `json:"ports"`
}
