package models

import (
	"go-deploy/pkg/subsystems/k8s/keys"
	"kubevirt.io/api/snapshot/v1alpha1"
	"time"
)

type VmSnapshotPublic struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Namespace string    `json:"namespace"`
	VmID      string    `json:"vmId"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

func (s *VmSnapshotPublic) Created() bool {
	return !s.CreatedAt.IsZero()
}

func (s *VmSnapshotPublic) IsPlaceholder() bool {
	return false
}

func CreateVmSnapshotPublicFromRead(vmSnapshot *v1alpha1.VirtualMachineSnapshot) *VmSnapshotPublic {
	var name string
	if vmSnapshot.ObjectMeta.Labels != nil {
		if n, ok := vmSnapshot.ObjectMeta.Labels[keys.LabelDeployName]; ok {
			name = n
		}
	}

	var status string
	if vmSnapshot.Status != nil {
		status = string(vmSnapshot.Status.Phase)
	} else {
		status = "Unknown"
	}

	return &VmSnapshotPublic{
		ID:        vmSnapshot.Name,
		Name:      name,
		Namespace: vmSnapshot.Namespace,
		VmID:      vmSnapshot.Spec.Source.Name,
		Status:    status,
		CreatedAt: formatCreatedAt(vmSnapshot.Annotations),
	}
}
