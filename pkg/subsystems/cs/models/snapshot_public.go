package models

import (
	"go-deploy/pkg/imp/cloudstack"
	"time"
)

type SnapshotPublic struct {
	ID          string    `json:"id"`
	VmID        string    `json:"vmId"`
	Name        string    `json:"displayname"`
	Description string    `json:"description"`
	ParentName  *string   `json:"parentName"`
	CreatedAt   time.Time `json:"created"`
	State       string    `json:"state"`
	Current     bool      `json:"current"`
}

func CreateSnapshotPublicFromGet(snapshot *cloudstack.VMSnapshot) *SnapshotPublic {
	iso8601 := "2006-01-02T15:04:05Z0700"
	createdAt, err := time.Parse(iso8601, snapshot.Created)
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
