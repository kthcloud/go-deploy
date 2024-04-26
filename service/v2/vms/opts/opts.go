package opts

import (
	"go-deploy/dto/v2/body"
	configModels "go-deploy/models/config"
	"go-deploy/models/model"
	"go-deploy/service/v2/utils"
	"time"
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
	UserID       string
	Zone         *configModels.Zone
	ExtraSshKeys []string
}

// GetOpts is used to specify the options when getting a VM.
type GetOpts struct {
	MigrationCode *string
	Shared        bool
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
	VmID          *string
	UserID        *string
	GpuGroupID    *string
	Pagination    *utils.Pagination
	CreatedBefore *time.Time
}

// GetGpuGroupOpts is used to specify the options when getting a GPU group.
type GetGpuGroupOpts struct {
}

// ListGpuGroupOpts is used to specify the options when listing GPU groups.
type ListGpuGroupOpts struct {
	Pagination *utils.Pagination
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
