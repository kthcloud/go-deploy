package api

import (
	"go-deploy/models/dto/v2/body"
	roleModels "go-deploy/models/sys/role"
	vmModels "go-deploy/models/sys/vm"
	"go-deploy/service/v2/vms/k8s_service"
	vmClient "go-deploy/service/v2/vms/opts"
)

type VMs interface {
	Get(id string, opts ...vmClient.GetOpts) (*vmModels.VM, error)
	List(opts ...vmClient.ListOpts) ([]vmModels.VM, error)
	Create(id, ownerID string, dtoVmCreate *body.VmCreate) error
	Update(id string, dtoVmUpdate *body.VmUpdate) error
	Delete(id string) error

	CheckQuota(id, userID string, quota *roleModels.Quotas, opts ...vmClient.QuotaOpts) error
	NameAvailable(name string) (bool, error)
	SshConnectionString(id string) (*string, error)

	GetSnapshot(vmID, id string, opts ...vmClient.GetSnapshotOpts) (*vmModels.SnapshotV2, error)
	GetSnapshotByName(vmID, name string, opts ...vmClient.GetSnapshotOpts) (*vmModels.SnapshotV2, error)
	ListSnapshots(vmID string, opts ...vmClient.ListSnapshotOpts) ([]vmModels.SnapshotV2, error)
	CreateSnapshot(vmID string, opts *vmClient.CreateSnapshotOpts) (*vmModels.SnapshotV2, error)
	DeleteSnapshot(vmID, id string) error

	K8s() *k8s_service.Client
}
