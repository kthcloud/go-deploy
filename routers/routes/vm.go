package routes

import "go-deploy/routers/api/v1/v1_vm"

const (
	VmsPath         = "/v1/vms"
	VmPath          = "/v1/vms/:vmId"
	VmSnapshotsPath = "/v1/vms/:vmId/snapshots"
	VmSnapshotPath  = "/v1/vms/:vmId/snapshots/:snapshotId"
	VmCommandPath   = "/v1/vms/:vmId/command"
)

type VmRoutingGroup struct{ RoutingGroupBase }

func VmRoutes() *VmRoutingGroup {
	return &VmRoutingGroup{}
}

func (group *VmRoutingGroup) PrivateRoutes() []Route {
	return []Route{
		{Method: "GET", Pattern: VmsPath, HandlerFunc: v1_vm.List},

		{Method: "GET", Pattern: VmPath, HandlerFunc: v1_vm.Get},
		{Method: "POST", Pattern: VmsPath, HandlerFunc: v1_vm.Create},
		{Method: "POST", Pattern: VmPath, HandlerFunc: v1_vm.Update},
		{Method: "DELETE", Pattern: VmPath, HandlerFunc: v1_vm.Delete},

		{Method: "GET", Pattern: VmSnapshotsPath, HandlerFunc: v1_vm.ListSnapshots},
		{Method: "GET", Pattern: VmSnapshotPath, HandlerFunc: v1_vm.GetSnapshot},
		{Method: "POST", Pattern: VmSnapshotsPath, HandlerFunc: v1_vm.CreateSnapshot},
		{Method: "DELETE", Pattern: VmSnapshotPath, HandlerFunc: v1_vm.DeleteSnapshot},

		{Method: "POST", Pattern: VmCommandPath, HandlerFunc: v1_vm.DoCommand},
	}
}
