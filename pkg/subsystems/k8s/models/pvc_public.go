package models

import (
	"go-deploy/pkg/subsystems/k8s/keys"
	v1 "k8s.io/api/core/v1"
)

type PvcPublic struct {
	ID        string `bson:"id"`
	Name      string `bson:"name"`
	Namespace string `bson:"namespace"`
	Capacity  string `bson:"capacity"`
	PvName    string `bson:"pvName"`
}

func (pvc *PvcPublic) Created() bool {
	return pvc.ID != ""
}

func CreatePvcPublicFromRead(pvc *v1.PersistentVolumeClaim) *PvcPublic {
	capacityQuantity, ok := pvc.Spec.Resources.Requests[v1.ResourceStorage]
	var capacity string
	if ok {
		capacity = capacityQuantity.String()
	}

	return &PvcPublic{
		ID:        pvc.Labels[keys.ManifestLabelID],
		Name:      pvc.Labels[keys.ManifestLabelName],
		Namespace: pvc.Namespace,
		Capacity:  capacity,
		PvName:    pvc.Spec.VolumeName,
	}
}