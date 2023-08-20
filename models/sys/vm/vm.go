package vm

import "time"

type VM struct {
	ID        string `bson:"id"`
	Name      string `bson:"name"`
	OwnerID   string `bson:"ownerId"`
	ManagedBy string `bson:"managedBy"`
	Zone      string `bson:"zone"`

	CreatedAt  time.Time `bson:"createdAt"`
	UpdatedAt  time.Time `bson:"updatedAt"`
	RepairedAt time.Time `bson:"repairedAt"`

	GpuID        string   `bson:"gpuId"`
	SshPublicKey string   `bson:"sshPublicKey"`
	Ports        []Port   `bson:"ports"`
	Activities   []string `bson:"activities"`

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
		if a == activity {
			return true
		}
	}
	return false
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
