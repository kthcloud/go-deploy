package vm

import (
	"fmt"
	"go-deploy/models/sys/activity"
	gpuModels "go-deploy/models/sys/gpu"
	"time"
)

type VM struct {
	ID      string `bson:"id"`
	Name    string `bson:"name"`
	Version string `bson:"version"`

	OwnerID   string `bson:"ownerId"`
	ManagedBy string `bson:"managedBy"`
	Host      *Host  `bson:"host,omitempty"`

	Zone string `bson:"zone"`
	// used for port http proxy, set in most cases, but kept as optional if no k8s is available
	DeploymentZone *string `bson:"deploymentZone"`

	CreatedAt  time.Time `bson:"createdAt"`
	UpdatedAt  time.Time `bson:"updatedAt,omitempty"`
	RepairedAt time.Time `bson:"repairedAt,omitempty"`
	DeletedAt  time.Time `bson:"deletedAt,omitempty"`

	NetworkID    string                       `bson:"networkId"`
	SshPublicKey string                       `bson:"sshPublicKey"`
	PortMap      map[string]Port              `bson:"portMap"`
	Activities   map[string]activity.Activity `bson:"activities"`

	Subsystems Subsystems `bson:"subsystems"`
	Specs      Specs      `bson:"specs"`

	StatusCode    int    `bson:"statusCode"`
	StatusMessage string `bson:"statusMessage"`

	Transfer *Transfer `bson:"transfer,omitempty"`
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
	exists, err := gpuModels.New().WithVM(vm.ID).ExistsAny()
	if err != nil {
		return false
	}

	return exists
}

func (vm *VM) GetGpuID() *string {
	idStruct, err := gpuModels.New().WithVM(vm.ID).GetID()
	if err != nil || idStruct == nil {
		return nil
	}

	return &idStruct.ID
}

func (vm *VM) GetGPU() *gpuModels.GPU {
	gpu, err := gpuModels.New().WithVM(vm.ID).Get()
	if err != nil || gpu == nil {
		return nil
	}

	return gpu
}

func (vm *VM) DoingOneOfActivities(activities []string) bool {
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

func (vm *VM) BeingTransferred() bool {
	return vm.Transfer != nil
}

func (vm *VM) GetExternalPort(privatePort int, protocol string) *int {
	pfrName := fmt.Sprintf("priv-%d-prot-%s", privatePort, protocol)
	service := vm.Subsystems.K8s.GetService(fmt.Sprintf("%s-%s", vm.Name, pfrName))
	if service == nil {
		return nil
	}

	for _, port := range service.Ports {
		if port.Name == pfrName {
			return &port.Port
		}
	}

	return nil
}
