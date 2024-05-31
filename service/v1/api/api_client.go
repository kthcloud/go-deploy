package api

import (
	"context"
	body2 "go-deploy/dto/v2/body"
	configModels "go-deploy/models/config"
	"go-deploy/models/model"
	"go-deploy/service/v1/deployments/harbor_service"
	"go-deploy/service/v1/deployments/k8s_service"
	dOpts "go-deploy/service/v1/deployments/opts"
	jobOpts "go-deploy/service/v1/jobs/opts"
	nOpts "go-deploy/service/v1/notifications/opts"
	resourceMigrationOpts "go-deploy/service/v1/resource_migrations/opts"
	smK8sService "go-deploy/service/v1/sms/k8s_service"
	smOpts "go-deploy/service/v1/sms/opts"
	statusOpts "go-deploy/service/v1/status/opts"
	teamOpts "go-deploy/service/v1/teams/opts"
	userOpts "go-deploy/service/v1/users/opts"
	zoneOpts "go-deploy/service/v1/zones/opts"
	"time"
)

type Deployments interface {
	Get(id string, opts ...dOpts.GetOpts) (*model.Deployment, error)
	GetByName(name string, opts ...dOpts.GetOpts) (*model.Deployment, error)
	List(opts ...dOpts.ListOpts) ([]model.Deployment, error)
	Create(id, userID string, dtoDeploymentCreate *body2.DeploymentCreate) error
	Update(id string, dtoDeploymentUpdate *body2.DeploymentUpdate) error
	UpdateOwner(id string, params *model.DeploymentUpdateOwnerParams) error
	Delete(id string) error
	Repair(id string) error

	Restart(id string) error
	DoCommand(id string, command string)

	GetCiConfig(id string) (*body2.CiConfig, error)

	SetupLogStream(id string, ctx context.Context, handler func(string, string, string, time.Time), history int) error
	AddLogs(id string, logs ...model.Log)

	StartActivity(id string, activity string) error
	CanAddActivity(id, activity string) (bool, string)

	CheckQuota(id string, params *dOpts.QuotaOptions) error
	NameAvailable(name string) (bool, error)
	GetUsage(userID string) (*model.DeploymentUsage, error)

	ValidateHarborToken(secret string) bool

	K8s() *k8s_service.Client
	Harbor() *harbor_service.Client
}

type Discovery interface {
	Discover() (*model.Discover, error)
}

type Events interface {
	Create(id string, params *model.EventCreateParams) error
}

type Jobs interface {
	Get(id string, opts ...jobOpts.GetOpts) (*model.Job, error)
	List(opts ...jobOpts.ListOpts) ([]model.Job, error)
	Create(id, userID, jobType, version string, args map[string]interface{}) error
	Update(id string, jobUpdateDTO *body2.JobUpdate) (*model.Job, error)
}

type Notifications interface {
	Get(id string, opts ...nOpts.GetOpts) (*model.Notification, error)
	List(opts ...nOpts.ListOpts) ([]model.Notification, error)
	Create(id, userID string, params *model.NotificationCreateParams) (*model.Notification, error)
	Update(id string, dtoNotificationUpdate *body2.NotificationUpdate) (*model.Notification, error)
	Delete(id string) error
}

type SMs interface {
	Get(id string, opts ...smOpts.GetOpts) (*model.SM, error)
	GetByUserID(userID string, opts ...smOpts.GetOpts) (*model.SM, error)
	List(opts ...smOpts.ListOpts) ([]model.SM, error)
	Create(id, userID string, params *model.SmCreateParams) error
	Delete(id string) error
	Repair(id string) error
	Exists(userID string) (bool, error)

	GetZone() *configModels.Zone
	GetUrlByUserID(userID string) *string

	K8s() *smK8sService.Client
}

type Status interface {
	ListWorkerStatus(opts ...statusOpts.ListWorkerStatusOpts) ([]model.WorkerStatus, error)
}

type Users interface {
	Get(id string, opts ...userOpts.GetOpts) (*model.User, error)
	GetByApiKey(apiKey string) (*model.User, error)
	GetUsage(userID string) (*model.UserUsage, error)
	List(opts ...userOpts.ListOpts) ([]model.User, error)
	ListTestUsers() ([]model.User, error)
	Synchronize(authParams *model.AuthParams) (*model.User, error)
	Update(userID string, dtoUserUpdate *body2.UserUpdate) (*model.User, error)
	Exists(id string) (bool, error)

	Discover(opts ...userOpts.DiscoverOpts) ([]body2.UserReadDiscovery, error)

	ApiKeys() ApiKeys
}

type ApiKeys interface {
	Create(userID string, dtoApiKeyCreate *body2.ApiKeyCreate) (*model.ApiKey, error)
}

type Teams interface {
	Get(id string, opts ...teamOpts.GetOpts) (*model.Team, error)
	List(opts ...teamOpts.ListOpts) ([]model.Team, error)
	ListIDs(opts ...teamOpts.ListOpts) ([]string, error)
	Create(id, ownerID string, dtoCreateTeam *body2.TeamCreate) (*model.Team, error)
	Update(id string, dtoUpdateTeam *body2.TeamUpdate) (*model.Team, error)
	Delete(id string) error
	CleanResource(id string) error
	Join(id string, dtoTeamJoin *body2.TeamJoin) (*model.Team, error)
	CheckResourceAccess(userID, resourceID string) (bool, error)
}

type Zones interface {
	Get(name string) *configModels.Zone
	List(opts ...zoneOpts.ListOpts) ([]configModels.Zone, error)
	HasCapability(zoneName, capability string) bool
}

type ResourceMigrations interface {
	Get(id string, opts ...resourceMigrationOpts.GetOpts) (*model.ResourceMigration, error)
	List(opts ...resourceMigrationOpts.ListOpts) ([]model.ResourceMigration, error)
	Create(id, userID string, migrationCreate *body2.ResourceMigrationCreate) (*model.ResourceMigration, *string, error)
	Update(id string, migrationUpdate *body2.ResourceMigrationUpdate, opts ...resourceMigrationOpts.UpdateOpts) (*model.ResourceMigration, *string, error)
	Delete(id string) error
}
