package opts

import v1 "github.com/kthcloud/go-deploy/service/v2/utils"

// GetOpts is used to specify the options when getting a resource migration.
type GetOpts struct {
	MigrationCode *string
}

// ListOpts is used to specify the options when listing resource migrations.
type ListOpts struct {
	Pagination *v1.Pagination
	UserID     *string
}

// UpdateOpts is used to specify the options when updating a resource migration.
type UpdateOpts struct {
	MigrationCode *string
}
