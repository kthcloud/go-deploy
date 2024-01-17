package vm

import (
	"go-deploy/models/sys/activity"
	gpuModels "go-deploy/models/sys/gpu"
	"time"
)

type VM struct {
	ID        string `bson:"id"`
	Name      string `bson:"name"`
	OwnerID   string `bson:"ownerId"`
	ManagedBy string `bson:"managedBy"`
	Host      *Host  `bson:"host,omitempty"`

	Zone string `bson:"zone"`
	// used for port http proxy, set in most cases, but kept as optional if no k8s is available
	DeploymentZone *string `bson:"deploymentZone"`

	CreatedAt  time.Time `bson:"createdAt"`
	UpdatedAt  time.Time `bson:"updatedAt"`
	RepairedAt time.Time `bson:"repairedAt"`
	DeletedAt  time.Time `bson:"deletedAt"`

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

func (vm *VM) GetGpu() *gpuModels.GPU {
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
