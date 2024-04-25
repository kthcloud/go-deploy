package opts

import v1 "go-deploy/service/v1/utils"

// GetOpts is used to specify the options when getting a resource migration.
type GetOpts struct {
}

// ListOpts is used to specify the options when listing resource migrations.
type ListOpts struct {
	Pagination *v1.Pagination
	UserID     *string
}
