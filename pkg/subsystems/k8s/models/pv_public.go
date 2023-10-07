package models

import (
	"go-deploy/pkg/subsystems/k8s/keys"
	v1 "k8s.io/api/core/v1"
	"time"
)

type PvPublic struct {
	ID        string    `bson:"id"`
	Name      string    `bson:"name"`
	Capacity  string    `bson:"capacity"`
	NfsServer string    `bson:"nfsServer"`
	NfsPath   string    `bson:"nfsPath"`
	CreatedAt time.Time `bson:"createdAt"`
}

func (pv *PvPublic) Created() bool {
	return pv.ID != ""
}

func (pv *PvPublic) IsPlaceholder() bool {
	return false
}

func CreatePvPublicFromRead(pv *v1.PersistentVolume) *PvPublic {
	capacityQuantity, ok := pv.Spec.Capacity[v1.ResourceStorage]
	var capacity string
	if ok {
		capacity = capacityQuantity.String()
	}

	var nfsServer string
	var nfsPath string
	if pv.Spec.NFS != nil {
		nfsServer = pv.Spec.NFS.Server
		nfsPath = pv.Spec.NFS.Path
	}

	return &PvPublic{
		ID:        pv.Labels[keys.ManifestLabelID],
		Name:      pv.Labels[keys.ManifestLabelName],
		Capacity:  capacity,
		NfsServer: nfsServer,
		NfsPath:   nfsPath,
		CreatedAt: formatCreatedAt(pv.Annotations),
	}
}
