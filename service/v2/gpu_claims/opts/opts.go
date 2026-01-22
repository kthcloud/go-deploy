package opts

import (
	configModels "github.com/kthcloud/go-deploy/models/config"
	"github.com/kthcloud/go-deploy/service/v2/utils"
)

// Opts is used to specify which resources to get.
type Opts struct {
	ClaimID   string
	Client    bool
	Generator bool

	ExtraOpts
}

// ExtraOpts is used to specify the extra options when getting a GpuClaim.
// This is useful when overwriting the implicit options
type ExtraOpts struct {
	Zone *configModels.Zone
}

// ListOpts is used to specify the options when listing gpu claim.
type ListOpts struct {
	Pagination *utils.Pagination
	All        bool
	Roles      *[]string
	Zone       *string
	Names      *[]string
}

// GetOpts is used to specify the options when getting a gpu claim.
type GetOpts struct{}
