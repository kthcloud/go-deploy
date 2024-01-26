package opts

import (
	"go-deploy/service/v1/utils"
)

// GetOpts is used to pass options to the Get method
type GetOpts struct {
}

// ListOpts is used to pass options to the List method
type ListOpts struct {
	Pagination *utils.Pagination
	UserID     *string
}
