package routes

import (
	v1 "go-deploy/routers/api/v1"
	v2 "go-deploy/routers/api/v2"
)

const (
	VmsPath         = "/v1/vms"
	VmPath          = "/v1/vms/:vmId"
	VmSnapshotsPath = "/v1/vms/:vmId/snapshots"
	VmSnapshotPath  = "/v1/vms/:vmId/snapshots/:snapshotId"
	VmCommandPath   = "/v1/vms/:vmId/command"

	VmsPathV2 = "/v2/vms"
	VmPathV2  = "/v2/vms/:vmId"
)

type VmRoutingGroup struct{ RoutingGroupBase }

func VmRoutes() *VmRoutingGroup {
	return &VmRoutingGroup{}
}

func (group *VmRoutingGroup) PrivateRoutes() []Route {
	return []Route{
		// V1
		{Method: "GET", Pattern: VmsPath, HandlerFunc: v1.ListVMs},

		{Method: "GET", Pattern: VmPath, HandlerFunc: v1.GetVM},
		// Deprecated
		//{Method: "POST", Pattern: VmsPath, HandlerFunc: v1.CreateVM},
		// Deprecated
		//{Method: "POST", Pattern: VmPath, HandlerFunc: v1.UpdateVM},
		{Method: "DELETE", Pattern: VmPath, HandlerFunc: v1.DeleteVM},

		{Method: "GET", Pattern: VmSnapshotsPath, HandlerFunc: v1.ListSnapshots},
		{Method: "GET", Pattern: VmSnapshotPath, HandlerFunc: v1.GetSnapshot},
		// Deprecated
		//{Method: "POST", Pattern: VmSnapshotsPath, HandlerFunc: v1.CreateSnapshot},
		{Method: "DELETE", Pattern: VmSnapshotPath, HandlerFunc: v1.DeleteSnapshot},

		{Method: "POST", Pattern: VmCommandPath, HandlerFunc: v1.DoVmCommand},

		// V2
		{Method: "GET", Pattern: VmPathV2, HandlerFunc: v2.GetVM},
		{Method: "GET", Pattern: VmsPathV2, HandlerFunc: v2.ListVMs},
		{Method: "POST", Pattern: VmsPathV2, HandlerFunc: v2.CreateVM},
		{Method: "POST", Pattern: VmPathV2, HandlerFunc: v2.UpdateVM},
		{Method: "DELETE", Pattern: VmPathV2, HandlerFunc: v2.DeleteVM},
	}
}
