package opts

import (
	"go-deploy/dto/v2/body"
	configModels "go-deploy/models/config"
	"go-deploy/models/model"
	"go-deploy/service/v2/utils"
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
	UserID string
	Zone   *configModels.DeploymentZone
}

// GetOpts is used to specify the options when getting a VM.
type GetOpts struct {
	TransferCode *string
	Shared       bool
}

// ListOpts is used to specify the options when listing VMs.
type ListOpts struct {
	Pagination *utils.Pagination
	UserID     *string
	Shared     bool
}

// GetGpuOpts is used to specify the options when getting a GPU.
type GetGpuOpts struct {
	Zone          *string
	AvailableGPUs bool
}

// ListGpuOpts is used to specify the options when listing GPUs.
type ListGpuOpts struct {
	Pagination    *utils.Pagination
	Zone          *string
	AvailableGPUs bool
}

// GetGpuLeaseOpts is used to specify the options when getting a GPU lease.
type GetGpuLeaseOpts struct {
}

// ListGpuLeaseOpts is used to specify the options when listing GPU leases.
type ListGpuLeaseOpts struct {
	Pagination *utils.Pagination
}

// CreateGpuLeaseOpts is used to specify the options when attaching a GPU to a VM.
type CreateGpuLeaseOpts struct {
	// LeaseForever is used to specify whether the lease should be created forever.
	LeaseForever bool
}

// GetSnapshotOpts is used to specify the options when getting a VM's snapshot.
type GetSnapshotOpts struct {
}

// ListSnapshotOpts is used to specify the options when listing VMs' snapshots.
type ListSnapshotOpts struct {
	Pagination *utils.Pagination
}

// CreateSnapshotOpts is used to specify the options when creating a VM's snapshot.
type CreateSnapshotOpts struct {
	System *model.CreateSnapshotParams
	User   *body.VmSnapshotCreate
}

// QuotaOpts is used to specify the options when getting a VM's quota.
type QuotaOpts struct {
	Quota          *model.Quotas
	Create         *body.VmCreate
	Update         *body.VmUpdate
	CreateSnapshot *body.VmSnapshotCreate
}
