package models

import (
	"go-deploy/pkg/imp/cloudstack"
	"strings"
	"time"
)

type SnapshotPublic struct {
	ID          string    `bson:"id"`
	VmID        string    `bson:"vmId"`
	Name        string    `bson:"name"`
	Description string    `bson:"description"`
	ParentName  *string   `bson:"parentName"`
	CreatedAt   time.Time `bson:"created"`
	State       string    `bson:"state"`
	Current     bool      `bson:"createdAt"`
}

func (s *SnapshotPublic) Created() bool {
	return s.ID != ""
}

func (s *SnapshotPublic) IsPlaceholder() bool {
	return false
}

func (s *SnapshotPublic) UserCreated() bool {
	return strings.Contains(s.Description, "user")
}

func (s *SnapshotPublic) SystemCreated() bool {
	return strings.Contains(s.Description, "system")
}

// CreateSnapshotPublicFromGet converts a cloudstack.VMSnapshot to a SnapshotPublic.
func CreateSnapshotPublicFromGet(snapshot *cloudstack.VMSnapshot) *SnapshotPublic {
	var parentName *string
	if snapshot.ParentName != "" {
		parentName = &snapshot.ParentName
	}

	return &SnapshotPublic{
		ID:          snapshot.Id,
		VmID:        snapshot.Virtualmachineid,
		Name:        snapshot.Displayname,
		Description: snapshot.Description,
		ParentName:  parentName,
		CreatedAt:   formatCreatedAt(snapshot.Created),
		State:       snapshot.State,
		Current:     snapshot.Current,
	}
}

// CreateSnapshotPublicFromCreate converts a cloudstack.CreateVMSnapshotResponse to a SnapshotPublic.
func CreateSnapshotPublicFromCreate(snapshot *cloudstack.CreateVMSnapshotResponse) *SnapshotPublic {
	var parentName *string
	if snapshot.ParentName != "" {
		parentName = &snapshot.ParentName
	}

	return &SnapshotPublic{
		ID:          snapshot.Id,
		VmID:        snapshot.Virtualmachineid,
		Name:        snapshot.Displayname,
		Description: snapshot.Description,
		ParentName:  parentName,
		CreatedAt:   formatCreatedAt(snapshot.Created),
		State:       snapshot.State,
		Current:     snapshot.Current,
	}
}
