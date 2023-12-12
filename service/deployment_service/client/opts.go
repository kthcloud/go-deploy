package client

import (
	configModels "go-deploy/models/config"
	"go-deploy/models/dto/body"
	roleModel "go-deploy/models/sys/role"
	"go-deploy/service"
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
	Zone   *configModels.DeploymentZone
}

// ListOptions is used to specify the options when listing deployments.
type ListOptions struct {
	Pagination      *service.Pagination
	GitHubWebhookID int64
	UserID          string
	Shared          bool
}

// GetOptions is used to specify the options when getting a deployment.
type GetOptions struct {
	TransferCode  string
	HarborWebhook *body.HarborWebhook
	Shared        bool
}

// QuotaOptions is used to specify the options when getting a deployment's quota.
type QuotaOptions struct {
	Quota  *roleModel.Quotas
	Create *body.DeploymentCreate
	Update *body.DeploymentUpdate
}
