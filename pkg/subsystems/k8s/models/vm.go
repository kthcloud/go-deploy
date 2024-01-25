package models

import (
	kubevirtv1 "kubevirt.io/api/core/v1"
	"time"
)

type VmPublic struct {
	Name      string `bson:"name"`
	Namespace string `bson:"namespace"`

	CpuCores int `bson:"cpu_cores"`
	RAM      int `bson:"memory"`
	DiskSize int `bson:"diskSize"`

	CloudInit string `bson:"cloudInit"`
	ImageURL  string `bson:"imageUrl"`
	PvName    string `bson:"pvName"`

	Running bool `bson:"running"`

	CreatedAt time.Time
}

func (vm *VmPublic) Created() bool {
	return !vm.CreatedAt.IsZero()
}

func (vm *VmPublic) IsPlaceholder() bool {
	return false
}

func CreateVmPublicFromRead(vm *kubevirtv1.VirtualMachine) *VmPublic {
	var running bool
	var ram int
	var cpuCores int
	var diskSize int
	var pvName string
	var cloudInit string
	var imageURL string

	if vm.Spec.Running != nil {
		running = *vm.Spec.Running
	}

	if vm.Spec.Template != nil {
		if r := vm.Spec.Template.Spec.Domain.Resources.Requests; r != nil {
			ram = int(r.Memory().Value())
			cpuCores = int(r.Cpu().Value())
		}

		for _, volume := range vm.Spec.Template.Spec.Volumes {
			if volume.CloudInitNoCloud != nil {
				cloudInit = volume.CloudInitNoCloud.UserData
			}
		}
	}

	if len(vm.Spec.DataVolumeTemplates) > 0 && vm.Spec.DataVolumeTemplates[0].Spec.PVC != nil {
		pvName = vm.Spec.DataVolumeTemplates[0].Spec.PVC.VolumeName
		if v := vm.Spec.DataVolumeTemplates[0].Spec.PVC.Resources.Requests; v != nil {
			diskSize = int(v.Storage().Value())
		}

		if vm.Spec.DataVolumeTemplates[0].Spec.Source != nil && vm.Spec.DataVolumeTemplates[0].Spec.Source.HTTP != nil {
			imageURL = vm.Spec.DataVolumeTemplates[0].Spec.Source.HTTP.URL
		}
	}

	return &VmPublic{
		Name:      vm.Name,
		Namespace: vm.Namespace,
		CpuCores:  cpuCores,
		RAM:       ram,
		DiskSize:  diskSize,
		CloudInit: cloudInit,
		ImageURL:  imageURL,
		PvName:    pvName,
		Running:   running,
		CreatedAt: formatCreatedAt(vm.Annotations),
	}
}
