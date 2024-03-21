package routes

import (
	"go-deploy/routers/api/v1/v1_vm"
	"go-deploy/routers/api/v2/v2_vm"
)

const (
	VmsPath         = "/v1/vms"
	VmPath          = "/v1/vms/:vmId"
	VmSnapshotsPath = "/v1/vms/:vmId/snapshots"
	VmSnapshotPath  = "/v1/vms/:vmId/snapshots/:snapshotId"
	VmCommandPath   = "/v1/vms/:vmId/command"

	VmsPathV2         = "/v2/vms"
	VmPathV2          = "/v2/vms/:vmId"
	VmGpuLeasesPathV2 = "/v2/vms/:vmId/gpuLease"
	VmGpuLeasePathV2  = "/v2/vms/:vmId/gpuLease/:gpuLeaseId"
	VmSnapshotsPathV2 = "/v2/vms/:vmId/snapshots"
	VmSnapshotPathV2  = "/v2/vms/:vmId/snapshots/:snapshotId"
	VmActionPathV2    = "/v2/vms/:vmId/action"
)

type VmRoutingGroup struct{ RoutingGroupBase }

func VmRoutes() *VmRoutingGroup {
	return &VmRoutingGroup{}
}

func (group *VmRoutingGroup) PrivateRoutes() []Route {
	return []Route{
		// V1
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

		// V2
		{Method: "GET", Pattern: VmsPathV2, HandlerFunc: v2_vm.List},
		{Method: "GET", Pattern: VmPathV2, HandlerFunc: v2_vm.Get},
		{Method: "POST", Pattern: VmsPathV2, HandlerFunc: v2_vm.Create},
		{Method: "POST", Pattern: VmPathV2, HandlerFunc: v2_vm.Update},
		{Method: "DELETE", Pattern: VmPathV2, HandlerFunc: v2_vm.Delete},

		// TODO: Implement GPU Leases
		//{Method: "GET", Pattern: VmGpuLeasesPathV2, HandlerFunc: v2_vm.ListGpuLeases},
		//{Method: "GET", Pattern: VmGpuLeasePathV2, HandlerFunc: v2_vm.GetGpuLease},
		//{Method: "POST", Pattern: VmGpuLeasesPathV2, HandlerFunc: v2_vm.CreateGpuLease},
		//{Method: "DELETE", Pattern: VmGpuLeasePathV2, HandlerFunc: v2_vm.DeleteGpuLease},

		{Method: "GET", Pattern: VmSnapshotsPathV2, HandlerFunc: v2_vm.ListSnapshots},
		{Method: "GET", Pattern: VmSnapshotPathV2, HandlerFunc: v2_vm.GetSnapshot},
		{Method: "POST", Pattern: VmSnapshotsPathV2, HandlerFunc: v2_vm.CreateSnapshot},
		{Method: "DELETE", Pattern: VmSnapshotPathV2, HandlerFunc: v2_vm.DeleteSnapshot},

		{Method: "POST", Pattern: VmActionPathV2, HandlerFunc: v2_vm.DoAction},
	}
}
