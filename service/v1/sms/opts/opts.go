package opts

import (
	configModels "go-deploy/models/config"
	"go-deploy/service/v1/common"
)

// Opts is used to specify which resources to get.
type Opts struct {
	SmID      string
	Client    bool
	Generator bool

	ExtraOpts
}

// ExtraOpts is used to specify the extra options when getting a VM.
// This is useful when overwriting the implicit options,
// such as where user ID is by default taken from StorageManager.OwnerID.
type ExtraOpts struct {
	UserID string
	Zone   *configModels.DeploymentZone
}

// ListOpts is used to specify the options when listing storage managers.
type ListOpts struct {
	Pagination *common.Pagination
	All        bool
}

// GetOpts is used to specify the options when getting a storage manager.
type GetOpts struct {
}
