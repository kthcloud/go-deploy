package client

import (
	"go-deploy/models/dto/body"
	roleModel "go-deploy/models/sys/role"
	vmModel "go-deploy/models/sys/vm"
	"go-deploy/service"
)

// Opts is used to specify which resources to get.
// For example, if you want to get only the VM, you can use OptsOnlyDeployment.
// If you want to get only the client, you can use OptsOnlyClient.
// If you want to get both the VM and the client, you can use OptsAll.
type Opts struct {
	VM        string
	Client    bool
	Generator bool
}

func OptsAll(id string) *Opts {
	return &Opts{
		VM:        id,
		Client:    true,
		Generator: true,
	}
}

func OptsNoVM() *Opts {
	return &Opts{
		VM:        "",
		Client:    true,
		Generator: true,
	}
}

func OptsNoClient(vmID string) *Opts {
	return &Opts{
		VM:        vmID,
		Client:    false,
		Generator: true,
	}
}

func OptsNoGenerator(vmID string) *Opts {
	return &Opts{
		VM:        vmID,
		Client:    true,
		Generator: false,
	}
}

func OptsOnlyClient() *Opts {
	return &Opts{
		VM:        "",
		Client:    true,
		Generator: false,
	}
}

// GetOptions is used to specify the options when getting a VM.
type GetOptions struct {
	TransferCode string
	Shared       bool
}

// ListOptions is used to specify the options when listing VMs.
type ListOptions struct {
	Pagination *service.Pagination
	UserID     string
	Shared     bool
}

// GetGpuOptions is used to specify the options when getting a VM's gpu.
type GetGpuOptions struct {
	AvailableGPUs bool
}

// ListGpuOptions is used to specify the options when listing VMs' gpus.
type ListGpuOptions struct {
	Pagination    *service.Pagination
	AvailableGPUs bool
}

// GetSnapshotOptions is used to specify the options when getting a VM's snapshot.
type GetSnapshotOptions struct {
}

// ListSnapshotOptions is used to specify the options when listing VMs' snapshots.
type ListSnapshotOptions struct {
	Pagination *service.Pagination
}

// CreateSnapshotOptions is used to specify the options when creating a VM's snapshot.
type CreateSnapshotOptions struct {
	System *vmModel.CreateSnapshotParams
	User   *body.VmSnapshotCreate
}

// QuotaOptions is used to specify the options when getting a VM's quota.
type QuotaOptions struct {
	Quota          *roleModel.Quotas
	Create         *body.VmCreate
	Update         *body.VmUpdate
	CreateSnapshot *body.VmSnapshotCreate
}
