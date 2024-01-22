package models

import (
	kubevirtv1 "kubevirt.io/api/core/v1"
	"time"
)

type VmPublic struct {
	Name      string `bson:"name"`
	Namespace string `bson:"namespace"`
	Running   bool   `bson:"running"`
	CreatedAt time.Time
}

func (vm *VmPublic) Created() bool {
	return !vm.CreatedAt.IsZero()
}

func (vm *VmPublic) IsPlaceholder() bool {
	return false
}

func CreateVmPublicFromRead(vm *kubevirtv1.VirtualMachine) *VmPublic {
	return &VmPublic{
		Name: vm.Name,
	}
}
