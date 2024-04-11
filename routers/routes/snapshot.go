package routes

import "go-deploy/routers/api/v2"

const (
	SnapshotsPath = "/v2/snapshots"
	SnapshotPath  = "/v2/snapshots/:snapshotId"
)

type SnapshotRoutingGroup struct{ RoutingGroupBase }

func SnapshotRoutes() *SnapshotRoutingGroup {
	return &SnapshotRoutingGroup{}
}

func (group *SnapshotRoutingGroup) PrivateRoutes() []Route {
	return []Route{
		{Method: "GET", Pattern: SnapshotsPath, HandlerFunc: v2.ListSnapshots},
		{Method: "GET", Pattern: SnapshotPath, HandlerFunc: v2.GetSnapshot},
		{Method: "POST", Pattern: SnapshotsPath, HandlerFunc: v2.CreateSnapshot},
		{Method: "DELETE", Pattern: SnapshotPath, HandlerFunc: v2.DeleteSnapshot},
	}
}
