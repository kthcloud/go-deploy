package models

import (
	v1 "k8s.io/api/core/v1"
	"time"
)

type PvPublic struct {
	Name      string `bson:"name"`
	Capacity  string `bson:"capacity"`
	NfsServer string `bson:"nfsServer"`
	NfsPath   string `bson:"nfsPath"`
	// Released is true if the volume is released.
	// This is mainly used to be able to repair the volume.
	// If it is released, then recreate the volume.
	Released  bool      `bson:"released"`
	CreatedAt time.Time `bson:"createdAt"`
}

func (pv *PvPublic) Created() bool {
	return pv.CreatedAt != time.Time{}
}

func (pv *PvPublic) IsPlaceholder() bool {
	return false
}

// CreatePvPublicFromRead creates a PvPublic from a v1.PersistentVolume.
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

	var released bool
	if pv.Status.Phase == v1.VolumeReleased {
		released = true
	}

	return &PvPublic{
		Name:      pv.Name,
		Capacity:  capacity,
		NfsServer: nfsServer,
		NfsPath:   nfsPath,
		Released:  released,
		CreatedAt: formatCreatedAt(pv.Annotations),
	}
}
