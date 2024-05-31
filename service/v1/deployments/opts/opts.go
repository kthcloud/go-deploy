package opts

import (
	body2 "go-deploy/dto/v2/body"
	configModels "go-deploy/models/config"
	v1 "go-deploy/service/v1/utils"
)

// Opts is used to specify which resources to get.
type Opts struct {
	DeploymentID string
	Client       bool
	Generator    bool

	ExtraOpts
}

// ExtraOpts is used to specify the extra options when getting a VM.
// This is useful when overwriting the implicit options,
// such as where user ID is by default taken from Deployment.OwnerID.
type ExtraOpts struct {
	UserID string
	Zone   *configModels.Zone
}

// ListOpts is used to specify the options when listing deployments.
type ListOpts struct {
	Pagination      *v1.Pagination
	GitHubWebhookID *int64
	UserID          *string
	Shared          bool
}

// GetOpts is used to specify the options when getting a deployment.
type GetOpts struct {
	MigrationCode *string
	HarborWebhook *body2.HarborWebhook
	Shared        bool
}

// QuotaOptions is used to specify the options when getting a deployment's quota.
type QuotaOptions struct {
	Create *body2.DeploymentCreate
	Update *body2.DeploymentUpdate
}
