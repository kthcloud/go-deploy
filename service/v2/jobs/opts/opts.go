package opts

import (
	"github.com/kthcloud/go-deploy/service/v2/utils"
)

// GetOpts is used to pass options to the Get method
type GetOpts struct {
}

// ListOpts is used to pass options to the List method
type ListOpts struct {
	Pagination *utils.Pagination
	SortBy     *utils.SortBy

	UserID          *string
	All             bool
	JobTypes        []string
	ExcludeJobTypes []string
	Status          []string
	ExcludeStatus   []string
}
