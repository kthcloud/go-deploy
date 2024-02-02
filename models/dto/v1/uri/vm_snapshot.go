package uri

type VmSnapshotCreate struct {
	VmID string `uri:"vmId" binding:"required,uuid4"`
}

type VmSnapshotList struct {
	VmID string `uri:"vmId" binding:"required,uuid4"`
}

type VmSnapshotGet struct {
	VmID       string `uri:"vmId" binding:"required,uuid4"`
	SnapshotID string `uri:"snapshotId" binding:"required,uuid4"`
}

type VmSnapshotDelete struct {
	VmID       string `uri:"vmId" binding:"required,uuid4"`
	SnapshotID string `uri:"snapshotId" binding:"required,uuid4"`
}
