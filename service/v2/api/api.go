package api

import (
	"go-deploy/dto/v2/body"
	"go-deploy/models/model"
	"go-deploy/service/v2/vms/k8s_service"
	vmOpts "go-deploy/service/v2/vms/opts"
)

type VMs interface {
	Get(id string, opts ...vmOpts.GetOpts) (*model.VM, error)
	List(opts ...vmOpts.ListOpts) ([]model.VM, error)
	Create(id, ownerID string, dtoVmCreate *body.VmCreate) error
	Update(id string, dtoVmUpdate *body.VmUpdate) error
	UpdateOwner(id string, params *model.VmUpdateOwnerParams) error
	Delete(id string) error
	Repair(id string) error

	IsAccessible(id string) (bool, error)

	CheckQuota(id, userID string, quota *model.Quotas, opts ...vmOpts.QuotaOpts) error
	NameAvailable(name string) (bool, error)
	SshConnectionString(id string) (*string, error)

	DoAction(id string, action *body.VmActionCreate) error

	Snapshots() Snapshots
	GPUs() GPUs
	GpuLeases() GpuLeases
	GpuGroups() GpuGroups

	K8s() *k8s_service.Client
}

type Snapshots interface {
	Get(vmID, id string, opts ...vmOpts.GetSnapshotOpts) (*model.SnapshotV2, error)
	GetByName(vmID, name string, opts ...vmOpts.GetSnapshotOpts) (*model.SnapshotV2, error)
	List(vmID string, opts ...vmOpts.ListSnapshotOpts) ([]model.SnapshotV2, error)
	Create(vmID string, opts ...vmOpts.CreateSnapshotOpts) (*model.SnapshotV2, error)
	Delete(vmID, id string) error
	Apply(vmID, id string) error
}

type GPUs interface {
}

type GpuLeases interface {
	Get(id string, opts ...vmOpts.GetGpuLeaseOpts) (*model.GpuLease, error)
	GetByVmID(vmID string, opts ...vmOpts.GetGpuLeaseOpts) (*model.GpuLease, error)
	List(opts ...vmOpts.ListGpuLeaseOpts) ([]model.GpuLease, error)
	Create(leaseID, userID string, dtoGpuLeaseCreate *body.GpuLeaseCreate) error
	Update(id string, dtoGpuLeaseUpdate *body.GpuLeaseUpdate) error
	Delete(id string) error

	Count(opts ...vmOpts.ListGpuLeaseOpts) (int, error)

	GetQueuePosition(id string) (int, error)
}

type GpuGroups interface {
	Get(id string, opts ...vmOpts.GetGpuGroupOpts) (*model.GpuGroup, error)
	List(opts ...vmOpts.ListGpuGroupOpts) ([]model.GpuGroup, error)
	Exists(id string) (bool, error)
}
