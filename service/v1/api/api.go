package api

import (
	"context"
	"go-deploy/dto/v1/body"
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
	"go-deploy/service/v1/vms/cs_service"
	vmK8sService "go-deploy/service/v1/vms/k8s_service"
	vmOpts "go-deploy/service/v1/vms/opts"
	zoneOpts "go-deploy/service/v1/zones/opts"
	"time"
)

type Deployments interface {
	Get(id string, opts ...dOpts.GetOpts) (*model.Deployment, error)
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
	List(opts ...userOpts.ListOpts) ([]model.User, error)
	Synchronize() (*model.User, error)
	Update(userID string, dtoUserUpdate *body.UserUpdate) (*model.User, error)
	Exists(id string) (bool, error)

	Discover(opts ...userOpts.DiscoverOpts) ([]body.UserReadDiscovery, error)
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

type VMs interface {
	Get(id string, opts ...vmOpts.GetOpts) (*model.VM, error)
	List(opts ...vmOpts.ListOpts) ([]model.VM, error)
	Create(id, ownerID string, dtoVmCreate *body.VmCreate) error
	Update(id string, dtoVmUpdate *body.VmUpdate) error
	Delete(id string) error
	Repair(id string) error

	GetSnapshot(vmID string, id string, opts ...vmOpts.GetSnapshotOpts) (*model.Snapshot, error)
	GetSnapshotByName(vmID string, name string, opts ...vmOpts.GetSnapshotOpts) (*model.Snapshot, error)
	ListSnapshots(vmID string, opts ...vmOpts.ListSnapshotOpts) ([]model.Snapshot, error)
	CreateSnapshot(id string, params *vmOpts.CreateSnapshotOpts) error
	DeleteSnapshot(id, snapshotID string) error
	ApplySnapshot(id, snapshotID string) error

	UpdateOwner(id string, params *body.VmUpdateOwner) error

	GetConnectionString(id string) (*string, error)
	GetExternalPortMapper(vmID string) (map[string]int, error)
	DoCommand(id, command string)

	GetHost(vmID string) (*model.Host, error)
	GetCloudStackHostCapabilities(hostName string, zoneName string) (*model.CloudStackHostCapabilities, error)

	StartActivity(id, activity string) error
	CanAddActivity(vmID, activity string) (bool, string, error)
	NameAvailable(name string) (bool, error)
	HttpProxyNameAvailable(id, name string) (bool, error)

	CheckQuota(id, userID string, quota *model.Quotas, opts ...vmOpts.QuotaOpts) error
	GetUsage(userID string) (*model.VmUsage, error)

	GetGPU(id string, opts ...vmOpts.GetGpuOpts) (*model.GPU, error)
	GetGpuByVM(vmID string) (*model.GPU, error)
	ListGPUs(opts ...vmOpts.ListGpuOpts) ([]model.GPU, error)
	AttachGPU(id string, gpuIDs []string, leaseDuration float64) error
	DetachGPU(id string) error
	IsGpuPrivileged(id string) (bool, error)
	CheckGpuHardwareAvailable(gpuID string) error
	CheckSuitableHost(id, hostName, zoneName string) error

	CS() *cs_service.Client
	K8s() *vmK8sService.Client
}

type Zones interface {
	Get(name string) *configModels.Zone
	GetLegacy(name string) *configModels.LegacyZone
	List(opts ...zoneOpts.ListOpts) ([]configModels.Zone, error)
	ListLegacy(opts ...zoneOpts.ListOpts) ([]configModels.LegacyZone, error)
	HasCapability(zoneName, capability string) bool
}

type ResourceMigrations interface {
	Get(id string, opts ...resourceMigrationOpts.GetOpts) (*model.ResourceMigration, error)
	List(opts ...resourceMigrationOpts.ListOpts) ([]model.ResourceMigration, error)
	Create(id, userID string, migrationCreate *body.ResourceMigrationCreate) (*model.ResourceMigration, *string, error)
	Update(id string, migrationUpdate *body.ResourceMigrationUpdate, opts ...resourceMigrationOpts.UpdateOpts) (*model.ResourceMigration, *string, error)
	Delete(id string) error
}
