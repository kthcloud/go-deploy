package api

import (
	"context"
	"go-deploy/dto/v2/body"
	configModels "go-deploy/models/config"
	"go-deploy/models/model"
	"go-deploy/service/v2/deployments/harbor_service"
	deploymentK8sService "go-deploy/service/v2/deployments/k8s_service"
	dOpts "go-deploy/service/v2/deployments/opts"
	jobOpts "go-deploy/service/v2/jobs/opts"
	nOpts "go-deploy/service/v2/notifications/opts"
	resourceMigrationOpts "go-deploy/service/v2/resource_migrations/opts"
	smK8sService "go-deploy/service/v2/sms/k8s_service"
	smOpts "go-deploy/service/v2/sms/opts"
	statusOpts "go-deploy/service/v2/status/opts"
	zoneOpts "go-deploy/service/v2/system/opts"
	teamOpts "go-deploy/service/v2/teams/opts"
	userOpts "go-deploy/service/v2/users/opts"
	vmK8sService "go-deploy/service/v2/vms/k8s_service"
	vmOpts "go-deploy/service/v2/vms/opts"
	"time"
)

type Deployments interface {
	Get(id string, opts ...dOpts.GetOpts) (*model.Deployment, error)
	GetByName(name string, opts ...dOpts.GetOpts) (*model.Deployment, error)
	List(opts ...dOpts.ListOpts) ([]model.Deployment, error)
	Create(id, userID string, dtoDeploymentCreate *body.DeploymentCreate) error
	Update(id string, dtoDeploymentUpdate *body.DeploymentUpdate) error
	UpdateOwner(id string, params *model.DeploymentUpdateOwnerParams) error
	Delete(id string) error
	Repair(id string) error

	Restart(id string) error
	DoCommand(id string, command string)

	GetCiConfig(id string) (*body.CiConfig, error)

	SetupLogStream(id string, ctx context.Context, handler func(string, string, string, time.Time), history int) error
	AddLogs(id string, logs ...model.Log)

	StartActivity(id string, activity string) error
	CanAddActivity(id, activity string) (bool, string)

	CheckQuota(id string, params *dOpts.QuotaOptions) error
	NameAvailable(name string) (bool, error)
	GetUsage(userID string) (*model.DeploymentUsage, error)

	ValidateHarborToken(secret string) bool

	K8s() *deploymentK8sService.Client
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
	Update(id string, jobUpdateDTO *body.JobUpdate) (*model.Job, error)
}

type Notifications interface {
	Get(id string, opts ...nOpts.GetOpts) (*model.Notification, error)
	List(opts ...nOpts.ListOpts) ([]model.Notification, error)
	Create(id, userID string, params *model.NotificationCreateParams) (*model.Notification, error)
	Update(id string, dtoNotificationUpdate *body.NotificationUpdate) (*model.Notification, error)
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
	Update(userID string, dtoUserUpdate *body.UserUpdate) (*model.User, error)
	Exists(id string) (bool, error)

	Discover(opts ...userOpts.DiscoverOpts) ([]body.UserReadDiscovery, error)

	ApiKeys() ApiKeys
}

type ApiKeys interface {
	Create(userID string, dtoApiKeyCreate *body.ApiKeyCreate) (*model.ApiKey, error)
}

type Teams interface {
	Get(id string, opts ...teamOpts.GetOpts) (*model.Team, error)
	List(opts ...teamOpts.ListOpts) ([]model.Team, error)
	ListIDs(opts ...teamOpts.ListOpts) ([]string, error)
	Create(id, ownerID string, dtoCreateTeam *body.TeamCreate) (*model.Team, error)
	Update(id string, dtoUpdateTeam *body.TeamUpdate) (*model.Team, error)
	Delete(id string) error
	CleanResource(id string) error
	Join(id string, dtoTeamJoin *body.TeamJoin) (*model.Team, error)
	CheckResourceAccess(userID, resourceID string) (bool, error)
}

type ResourceMigrations interface {
	Get(id string, opts ...resourceMigrationOpts.GetOpts) (*model.ResourceMigration, error)
	List(opts ...resourceMigrationOpts.ListOpts) ([]model.ResourceMigration, error)
	Create(id, userID string, migrationCreate *body.ResourceMigrationCreate) (*model.ResourceMigration, *string, error)
	Update(id string, migrationUpdate *body.ResourceMigrationUpdate, opts ...resourceMigrationOpts.UpdateOpts) (*model.ResourceMigration, *string, error)
	Delete(id string) error
}

type VMs interface {
	Get(id string, opts ...vmOpts.GetOpts) (*model.VM, error)
	List(opts ...vmOpts.ListOpts) ([]model.VM, error)
	Create(id, ownerID string, dtoVmCreate *body.VmCreate) error
	Update(id string, dtoVmUpdate *body.VmUpdate) error
	UpdateOwner(id string, params *model.VmUpdateOwnerParams) error
	Delete(id string) error
	Repair(id string) error

	IsAccessible(id string) (bool, error)

	CheckQuota(id, userID string, quota *model.Quotas, opts ...vmOpts.QuotaOpts) error
	GetUsage(userID string) (*model.VmUsage, error)
	NameAvailable(name string) (bool, error)
	SshConnectionString(id string) (*string, error)

	DoAction(id string, action *body.VmActionCreate) error

	Snapshots() Snapshots
	GpuLeases() GpuLeases
	GpuGroups() GpuGroups

	K8s() *vmK8sService.Client
}

type Snapshots interface {
	Get(vmID, id string, opts ...vmOpts.GetSnapshotOpts) (*model.SnapshotV2, error)
	GetByName(vmID, name string, opts ...vmOpts.GetSnapshotOpts) (*model.SnapshotV2, error)
	List(vmID string, opts ...vmOpts.ListSnapshotOpts) ([]model.SnapshotV2, error)
	Create(vmID string, opts ...vmOpts.CreateSnapshotOpts) (*model.SnapshotV2, error)
	Delete(vmID, id string) error
	Apply(vmID, id string) error
}

type GPUs interface {
}

type GpuLeases interface {
	Get(id string, opts ...vmOpts.GetGpuLeaseOpts) (*model.GpuLease, error)
	GetByVmID(vmID string, opts ...vmOpts.GetGpuLeaseOpts) (*model.GpuLease, error)
	List(opts ...vmOpts.ListGpuLeaseOpts) ([]model.GpuLease, error)
	Create(leaseID, userID string, dtoGpuLeaseCreate *body.GpuLeaseCreate) error
	Update(id string, dtoGpuLeaseUpdate *body.GpuLeaseUpdate) error
	Delete(id string) error

	Count(opts ...vmOpts.ListGpuLeaseOpts) (int, error)

	GetQueuePosition(id string) (int, error)
}

type GpuGroups interface {
	Get(id string, opts ...vmOpts.GetGpuGroupOpts) (*model.GpuGroup, error)
	List(opts ...vmOpts.ListGpuGroupOpts) ([]model.GpuGroup, error)
	Exists(id string) (bool, error)
}

type System interface {
	ListCapacities(n int) ([]body.TimestampedSystemCapacities, error)
	ListStats(n int) ([]body.TimestampedSystemStats, error)
	ListStatus(n int) ([]body.TimestampedSystemStatus, error)

	RegisterNode(params *body.HostRegisterParams) error

	ListHosts() ([]model.Host, error)

	GetZone(name string) *configModels.Zone
	ListZones(opts ...zoneOpts.ListOpts) ([]configModels.Zone, error)
	ZoneHasCapability(zoneName, capability string) bool
}
