package models

import (
	"go-deploy/pkg/imp/cloudstack"
	"time"
)

type SnapshotPublic struct {
	ID         string    `json:"id"`
	VmID       string    `json:"vmId"`
	Name       string    `json:"displayname"`
	ParentName *string   `json:"parentName"`
	CreatedAt  time.Time `json:"created"`
	State      string    `json:"state"`
	Current    bool      `json:"current"`
}

func CreateSnapshotPublicFromGet(snapshot *cloudstack.VMSnapshot) *SnapshotPublic {
	createdAt, err := time.Parse(time.RFC3339, snapshot.Created)
	if err != nil {
		createdAt = time.Now()
	}

	var parentName *string
	if snapshot.ParentName != "" {
		parentName = &snapshot.ParentName
	}

	return &SnapshotPublic{
		ID:         snapshot.Id,
		VmID:       snapshot.Virtualmachineid,
		Name:       snapshot.Displayname,
		ParentName: parentName,
		CreatedAt:  createdAt,
		State:      snapshot.State,
		Current:    snapshot.Current,
	}
}
