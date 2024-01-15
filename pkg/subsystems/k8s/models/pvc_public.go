package models

import (
	"go-deploy/pkg/subsystems/k8s/keys"
	v1 "k8s.io/api/core/v1"
	"time"
)

type PvcPublic struct {
	ID        string    `bson:"id"`
	Name      string    `bson:"name"`
	Namespace string    `bson:"namespace"`
	Capacity  string    `bson:"capacity"`
	PvName    string    `bson:"pvName"`
	CreatedAt time.Time `bson:"createdAt"`
}

func (pvc *PvcPublic) Created() bool {
	return pvc.ID != ""
}

func (pvc *PvcPublic) IsPlaceholder() bool {
	return false
}

// CreatePvcPublicFromRead creates a PvcPublic from a v1.PersistentVolumeClaim.
func CreatePvcPublicFromRead(pvc *v1.PersistentVolumeClaim) *PvcPublic {
	capacityQuantity, ok := pvc.Spec.Resources.Requests[v1.ResourceStorage]
	var capacity string
	if ok {
		capacity = capacityQuantity.String()
	}

	return &PvcPublic{
		ID:        pvc.Labels[keys.ManifestLabelID],
		Name:      pvc.Name,
		Namespace: pvc.Namespace,
		Capacity:  capacity,
		PvName:    pvc.Spec.VolumeName,
		CreatedAt: formatCreatedAt(pvc.Annotations),
	}
}
