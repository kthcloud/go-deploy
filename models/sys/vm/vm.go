package vm

import (
	"go-deploy/models/sys/activity"
	"time"
)

type VM struct {
	ID        string `bson:"id"`
	Name      string `bson:"name"`
	OwnerID   string `bson:"ownerId"`
	ManagedBy string `bson:"managedBy"`

	Zone string `bson:"zone"`
	// used for port http proxy, set in most cases, but kept as optional if no k8s is available
	DeploymentZone *string `bson:"deploymentZone"`

	CreatedAt  time.Time `bson:"createdAt"`
	UpdatedAt  time.Time `bson:"updatedAt"`
	RepairedAt time.Time `bson:"repairedAt"`
	DeletedAt  time.Time `bson:"deletedAt"`

	NetworkID    string                       `bson:"networkId"`
	GpuID        string                       `bson:"gpuId"`
	SshPublicKey string                       `bson:"sshPublicKey"`
	Ports        []Port                       `bson:"ports"`
	Activities   map[string]activity.Activity `bson:"activities"`

	Subsystems Subsystems `bson:"subsystems"`
	Specs      Specs      `bson:"specs"`

	StatusCode    int    `bson:"statusCode"`
	StatusMessage string `bson:"statusMessage"`
}

type Specs struct {
	CpuCores int `json:"cpuCores"`
	RAM      int `json:"ram"`
	DiskSize int `json:"diskSize"`
}

func (vm *VM) Ready() bool {
	return !vm.DoingActivity(ActivityBeingCreated) && !vm.DoingActivity(ActivityBeingDeleted)
}

func (vm *VM) DoingActivity(activity string) bool {
	for _, a := range vm.Activities {
		if a.Name == activity {
			return true
		}
	}
	return false
}

func (vm *VM) HasGPU() bool {
	return vm.GpuID != ""
}

func (vm *VM) DoingOnOfActivities(activities []string) bool {
	for _, a := range activities {
		if vm.DoingActivity(a) {
			return true
		}
	}
	return false
}

func (vm *VM) BeingCreated() bool {
	return vm.DoingActivity(ActivityBeingCreated)
}

func (vm *VM) BeingDeleted() bool {
	return vm.DoingActivity(ActivityBeingDeleted)
}
