package models

import (
	"go-deploy/pkg/subsystems/k8s/keys"
	kubevirtv1 "kubevirt.io/api/core/v1"
	"time"
)

type VmPublic struct {
	ID        string            `bson:"id"`
	Name      string            `bson:"name"`
	Namespace string            `bson:"namespace"`
	Labels    map[string]string `bson:"labels"`

	CpuCores int      `bson:"cpuCores"`
	RAM      int      `bson:"memory"`
	DiskSize int      `bson:"diskSize"`
	GPUs     []string `bson:"gpus"`

	CloudInit string `bson:"cloudInit"`
	// Image is the URL of the image to use for the VM
	// It may either be an HTTP URL or a Docker image.
	//
	// If it is an HTTP URL, it must be in the format: http(s)://<url>
	// If it is a Docker image, it must be in the format: docker://<image>
	Image string `bson:"image"`

	Running bool `bson:"running"`

	CreatedAt time.Time `bson:"createdAt"`
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
	var cloudInit string
	var image string
	var name string

	if vm.ObjectMeta.Labels != nil {
		if n, ok := vm.ObjectMeta.Labels[keys.LabelDeployName]; ok {
			name = n
		}
	}

	if vm.Spec.Running != nil {
		running = *vm.Spec.Running
	}

	if vm.Spec.Template != nil {
		if r := vm.Spec.Template.Spec.Domain.Resources.Limits; r != nil {
			ram = int(r.Memory().Value()) / 1024 / 1024 / 1024
			cpuCores = int(r.Cpu().Value())
		}

		for _, volume := range vm.Spec.Template.Spec.Volumes {
			if volume.CloudInitNoCloud != nil {
				cloudInit = volume.CloudInitNoCloud.UserData
			}
		}
	}

	if len(vm.Spec.DataVolumeTemplates) > 0 && vm.Spec.DataVolumeTemplates[0].Spec.PVC != nil {
		if v := vm.Spec.DataVolumeTemplates[0].Spec.PVC.Resources.Requests; v != nil {
			diskSize = int(v.Storage().Value()) / 1024 / 1024 / 1024
		}

		source := vm.Spec.DataVolumeTemplates[0].Spec.Source
		if source != nil {
			if source.Registry != nil && source.Registry.URL != nil {
				image = *source.Registry.URL
			} else if source.HTTP != nil {
				image = source.HTTP.URL
			}
		}
	}

	return &VmPublic{
		ID:        vm.Name,
		Name:      name,
		Namespace: vm.Namespace,
		CpuCores:  cpuCores,
		RAM:       ram,
		DiskSize:  diskSize,
		CloudInit: cloudInit,
		Image:     image,
		Running:   running,
		CreatedAt: formatCreatedAt(vm.Annotations),
	}
}
