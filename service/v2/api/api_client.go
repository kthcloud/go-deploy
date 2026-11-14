package api

import (
	"context"
	"time"

	body "github.com/kthcloud/go-deploy/api/openapi/v2"
	configModels "github.com/kthcloud/go-deploy/models/config"
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/service/v2/deployments/harbor_service"
	deploymentK8sService "github.com/kthcloud/go-deploy/service/v2/deployments/k8s_service"
	dOpts "github.com/kthcloud/go-deploy/service/v2/deployments/opts"
	gpuClaimsK8sService "github.com/kthcloud/go-deploy/service/v2/gpu_claims/k8s_service"
	gpuClaimOpts "github.com/kthcloud/go-deploy/service/v2/gpu_claims/opts"
	jobOpts "github.com/kthcloud/go-deploy/service/v2/jobs/opts"
	nOpts "github.com/kthcloud/go-deploy/service/v2/notifications/opts"
	resourceMigrationOpts "github.com/kthcloud/go-deploy/service/v2/resource_migrations/opts"
	smK8sService "github.com/kthcloud/go-deploy/service/v2/sms/k8s_service"
	smOpts "github.com/kthcloud/go-deploy/service/v2/sms/opts"
	systemOpts "github.com/kthcloud/go-deploy/service/v2/system/opts"
	teamOpts "github.com/kthcloud/go-deploy/service/v2/teams/opts"
	userOpts "github.com/kthcloud/go-deploy/service/v2/users/opts"
	vmK8sService "github.com/kthcloud/go-deploy/service/v2/vms/k8s_service"
	vmOpts "github.com/kthcloud/go-deploy/service/v2/vms/opts"
)

type Deployments interface {
	Get(ctx context.Context, id string, opts ...dOpts.GetOpts) (*model.Deployment, error)
	GetByName(ctx context.Context, name string, opts ...dOpts.GetOpts) (*model.Deployment, error)
	List(ctx context.Context, opts ...dOpts.ListOpts) ([]model.Deployment, error)
	Create(ctx context.Context, id, userID string, dtoDeploymentCreate *body.DeploymentCreate) error
	Update(ctx context.Context, id string, dtoDeploymentUpdate *body.DeploymentUpdate) error
	UpdateOwner(ctx context.Context, id string, params *model.DeploymentUpdateOwnerParams) error
	Delete(ctx context.Context, id string) error
	Repair(ctx context.Context, id string) error

	Restart(ctx context.Context, id string) error
	DoCommand(ctx context.Context, id string, command string)

	GetCiConfig(ctx context.Context, id string) (*body.CiConfig, error)

	SetupLogStream(ctx context.Context, id string, handler func(string, string, string, time.Time), history int) error
	AddLogs(ctx context.Context, id string, logs ...model.Log)

	StartActivity(ctx context.Context, id string, activity string) error
	CanAddActivity(ctx context.Context, id, activity string) (bool, string)

	CheckQuota(ctx context.Context, id string, params *dOpts.QuotaOptions) error
	NameAvailable(ctx context.Context, name string) (bool, error)
	GetUsage(ctx context.Context, userID string) (*model.DeploymentUsage, error)

	ValidateHarborToken(secret string) bool

	K8s() *deploymentK8sService.Client
	Harbor() *harbor_service.Client
}

type Discovery interface {
	Discover(ctx context.Context) (*model.Discover, error)
}

type Events interface {
	Create(ctx context.Context, id string, params *model.EventCreateParams) error
}

type Jobs interface {
	Get(ctx context.Context, id string, opts ...jobOpts.GetOpts) (*model.Job, error)
	List(ctx context.Context, opts ...jobOpts.ListOpts) ([]model.Job, error)
	Create(ctx context.Context, id, userID, jobType, version string, args map[string]any) error
	Update(ctx context.Context, id string, jobUpdateDTO *body.JobUpdate) (*model.Job, error)
}

type Notifications interface {
	Get(ctx context.Context, id string, opts ...nOpts.GetOpts) (*model.Notification, error)
	List(ctx context.Context, opts ...nOpts.ListOpts) ([]model.Notification, error)
	Create(ctx context.Context, id, userID string, params *model.NotificationCreateParams) (*model.Notification, error)
	Update(ctx context.Context, id string, dtoNotificationUpdate *body.NotificationUpdate) (*model.Notification, error)
	Delete(ctx context.Context, id string) error
}

type SMs interface {
	Get(ctx context.Context, id string, opts ...smOpts.GetOpts) (*model.SM, error)
	GetByUserID(ctx context.Context, userID string, opts ...smOpts.GetOpts) (*model.SM, error)
	List(ctx context.Context, opts ...smOpts.ListOpts) ([]model.SM, error)
	Create(ctx context.Context, id, userID string, params *model.SmCreateParams) error
	Delete(ctx context.Context, id string) error
	Repair(ctx context.Context, id string) error
	Exists(ctx context.Context, userID string) (bool, error)

	GetZone() *configModels.Zone
	GetUrlByUserID(ctx context.Context, userID string) *string

	K8s() *smK8sService.Client
}

type Users interface {
	Get(ctx context.Context, id string, opts ...userOpts.GetOpts) (*model.User, error)
	GetByApiKey(ctx context.Context, apiKey string) (*model.User, error)
	GetUsage(ctx context.Context, userID string) (*model.UserUsage, error)
	List(ctx context.Context, opts ...userOpts.ListOpts) ([]model.User, error)
	ListTestUsers(ctx context.Context) ([]model.User, error)
	Synchronize(ctx context.Context, authParams *model.AuthParams) (*model.User, error)
	Update(ctx context.Context, userID string, dtoUserUpdate *body.UserUpdate) (*model.User, error)
	Exists(ctx context.Context, id string) (bool, error)

	Discover(ctx context.Context, opts ...userOpts.DiscoverOpts) ([]body.UserReadDiscovery, error)

	ApiKeys() ApiKeys
}

type ApiKeys interface {
	Create(ctx context.Context, userID string, dtoApiKeyCreate *body.ApiKeyCreate) (*model.ApiKey, error)
}

type Teams interface {
	Get(ctx context.Context, id string, opts ...teamOpts.GetOpts) (*model.Team, error)
	List(ctx context.Context, opts ...teamOpts.ListOpts) ([]model.Team, error)
	ListIDs(ctx context.Context, opts ...teamOpts.ListOpts) ([]string, error)
	Create(ctx context.Context, id, ownerID string, dtoCreateTeam *body.TeamCreate) (*model.Team, error)
	Update(ctx context.Context, id string, dtoUpdateTeam *body.TeamUpdate) (*model.Team, error)
	Delete(ctx context.Context, id string) error
	CleanResource(ctx context.Context, id string) error
	Join(ctx context.Context, id string, dtoTeamJoin *body.TeamJoin) (*model.Team, error)
	CheckResourceAccess(ctx context.Context, userID, resourceID string) (bool, error)
}

type ResourceMigrations interface {
	Get(ctx context.Context, id string, opts ...resourceMigrationOpts.GetOpts) (*model.ResourceMigration, error)
	List(ctx context.Context, opts ...resourceMigrationOpts.ListOpts) ([]model.ResourceMigration, error)
	Create(ctx context.Context, id, userID string, migrationCreate *body.ResourceMigrationCreate) (*model.ResourceMigration, *string, error)
	Update(ctx context.Context, id string, migrationUpdate *body.ResourceMigrationUpdate, opts ...resourceMigrationOpts.UpdateOpts) (*model.ResourceMigration, *string, error)
	Delete(ctx context.Context, id string) error
}

type VMs interface {
	Get(ctx context.Context, id string, opts ...vmOpts.GetOpts) (*model.VM, error)
	List(ctx context.Context, opts ...vmOpts.ListOpts) ([]model.VM, error)
	Create(ctx context.Context, id, ownerID string, dtoVmCreate *body.VmCreate) error
	Update(ctx context.Context, id string, dtoVmUpdate *body.VmUpdate) error
	UpdateOwner(ctx context.Context, id string, params *model.VmUpdateOwnerParams) error
	Delete(ctx context.Context, id string) error
	Repair(ctx context.Context, id string) error

	IsAccessible(ctx context.Context, id string) (bool, error)

	CheckQuota(ctx context.Context, id, userID string, quota *model.Quotas, opts ...vmOpts.QuotaOpts) error
	GetUsage(ctx context.Context, userID string) (*model.VmUsage, error)
	NameAvailable(ctx context.Context, name string) (bool, error)
	SshConnectionString(ctx context.Context, id string) (*string, error)

	DoAction(ctx context.Context, id string, action *body.VmActionCreate) error

	Snapshots() Snapshots
	GpuLeases() GpuLeases
	GpuGroups() GpuGroups

	K8s() *vmK8sService.Client
}

type Snapshots interface {
	Get(ctx context.Context, vmID, id string, opts ...vmOpts.GetSnapshotOpts) (*model.SnapshotV2, error)
	GetByName(ctx context.Context, vmID, name string, opts ...vmOpts.GetSnapshotOpts) (*model.SnapshotV2, error)
	List(ctx context.Context, vmID string, opts ...vmOpts.ListSnapshotOpts) ([]model.SnapshotV2, error)
	Create(ctx context.Context, vmID string, opts ...vmOpts.CreateSnapshotOpts) (*model.SnapshotV2, error)
	Delete(ctx context.Context, vmID, id string) error
	Apply(ctx context.Context, vmID, id string) error
}

type GPUs any

type GpuLeases interface {
	Get(ctx context.Context, id string, opts ...vmOpts.GetGpuLeaseOpts) (*model.GpuLease, error)
	GetByVmID(ctx context.Context, vmID string, opts ...vmOpts.GetGpuLeaseOpts) (*model.GpuLease, error)
	List(ctx context.Context, opts ...vmOpts.ListGpuLeaseOpts) ([]model.GpuLease, error)
	Create(ctx context.Context, leaseID, userID string, dtoGpuLeaseCreate *body.GpuLeaseCreate) error
	Update(ctx context.Context, id string, dtoGpuLeaseUpdate *body.GpuLeaseUpdate) error
	Delete(ctx context.Context, id string) error

	Count(ctx context.Context, opts ...vmOpts.ListGpuLeaseOpts) (int, error)

	GetQueuePosition(ctx context.Context, id string) (int, error)
}

type GpuGroups interface {
	Get(ctx context.Context, id string, opts ...vmOpts.GetGpuGroupOpts) (*model.GpuGroup, error)
	List(ctx context.Context, opts ...vmOpts.ListGpuGroupOpts) ([]model.GpuGroup, error)
	Exists(ctx context.Context, id string) (bool, error)
}

type System interface {
	ListCapacities(ctx context.Context, n int) ([]body.TimestampedSystemCapacities, error)
	ListStats(ctx context.Context, n int) ([]body.TimestampedSystemStats, error)
	ListStatus(ctx context.Context, n int) ([]body.TimestampedSystemStatus, error)

	ListWorkerStatus(ctx context.Context, opts ...systemOpts.ListWorkerStatusOpts) ([]model.WorkerStatus, error)

	RegisterNode(ctx context.Context, params *body.HostRegisterParams) error

	ListHosts(ctx context.Context) ([]model.Host, error)
	ListAllHosts(ctx context.Context) ([]model.Host, error)

	GetZone(name string) *configModels.Zone
	ListZones(opts ...systemOpts.ListOpts) ([]configModels.Zone, error)
	ZoneHasCapability(zoneName, capability string) bool
}

type GpuClaims interface {
	Get(ctx context.Context, id string, opts ...gpuClaimOpts.Opts) (*model.GpuClaim, error)
	List(ctx context.Context, opts ...gpuClaimOpts.ListOpts) ([]model.GpuClaim, error)
	Create(ctx context.Context, id string, gpuClaimsCreateParams *model.GpuClaimCreateParams) error
	Delete(ctx context.Context, id string) error
	Update(ctx context.Context, id string, gpuClaimsUpdateParams *model.GpuClaimUpdateParams) error

	K8s() *gpuClaimsK8sService.Client
}
