package client

import (
	configModels "go-deploy/models/config"
	"go-deploy/models/dto/body"
	roleModels "go-deploy/models/sys/role"
	vmModels "go-deploy/models/sys/vm"
	"go-deploy/service"
)

// Opts is used to specify which resources to get.
type Opts struct {
	VmID      string
	Client    bool
	Generator bool

	ExtraOpts
}

// ExtraOpts is used to specify the extra options when getting a VM.
// This is useful when overwriting the implicit options,
// such as where user ID is by default taken from VM.OwnerID.
type ExtraOpts struct {
	UserID         string
	Zone           *configModels.VmZone
	DeploymentZone *configModels.DeploymentZone
}

// GetOptions is used to specify the options when getting a VM.
type GetOptions struct {
	TransferCode *string
	Shared       bool
}

// ListOptions is used to specify the options when listing VMs.
type ListOptions struct {
	Pagination *service.Pagination
	UserID     *string
	Shared     bool
}

// GetGpuOptions is used to specify the options when getting a VM's gpu.
type GetGpuOptions struct {
	Zone          *string
	AvailableGPUs bool
}

// ListGpuOptions is used to specify the options when listing VMs' gpus.
type ListGpuOptions struct {
	Pagination    *service.Pagination
	Zone          *string
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
	System *vmModels.CreateSnapshotParams
	User   *body.VmSnapshotCreate
}

// QuotaOptions is used to specify the options when getting a VM's quota.
type QuotaOptions struct {
	Quota          *roleModels.Quotas
	Create         *body.VmCreate
	Update         *body.VmUpdate
	CreateSnapshot *body.VmSnapshotCreate
}
