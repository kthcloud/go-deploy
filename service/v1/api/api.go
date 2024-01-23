package api

import (
	"context"
	configModels "go-deploy/models/config"
	"go-deploy/models/dto/v1/body"
	deploymentModels "go-deploy/models/sys/deployment"
	"go-deploy/models/sys/discover"
	eventModels "go-deploy/models/sys/event"
	gpuModels "go-deploy/models/sys/gpu"
	jobModels "go-deploy/models/sys/job"
	notificationModels "go-deploy/models/sys/notification"
	roleModels "go-deploy/models/sys/role"
	smModels "go-deploy/models/sys/sm"
	teamModels "go-deploy/models/sys/team"
	userModels "go-deploy/models/sys/user"
	vmModels "go-deploy/models/sys/vm"
	workerStatusModels "go-deploy/models/sys/worker_status"
	zoneModels "go-deploy/models/sys/zone"
	"go-deploy/service/v1/deployments/github_service"
	"go-deploy/service/v1/deployments/harbor_service"
	"go-deploy/service/v1/deployments/k8s_service"
	dOpts "go-deploy/service/v1/deployments/opts"
	"go-deploy/service/v1/jobs/opts"
	opts2 "go-deploy/service/v1/notifications/opts"
	sK8sService "go-deploy/service/v1/sms/k8s_service"
	smClient "go-deploy/service/v1/sms/opts"
	opts3 "go-deploy/service/v1/status/opts"
	opts4 "go-deploy/service/v1/teams/opts"
	opts5 "go-deploy/service/v1/users/opts"
	"go-deploy/service/v1/vms/cs_service"
	vK8sService "go-deploy/service/v1/vms/k8s_service"
	vmClient "go-deploy/service/v1/vms/opts"
	opts6 "go-deploy/service/v1/zones/opts"
)

type Deployments interface {
	Get(id string, opts ...dOpts.GetOpts) (*deploymentModels.Deployment, error)
	List(opts ...dOpts.ListOpts) ([]deploymentModels.Deployment, error)
	Create(id, userID string, dtoDeploymentCreate *body.DeploymentCreate) error
	Update(id string, dtoDeploymentUpdate *body.DeploymentUpdate) error
	Delete(id string) error
	Repair(id string) error

	UpdateOwnerSetup(id string, params *body.DeploymentUpdateOwner) (*string, error)
	UpdateOwner(id string, params *body.DeploymentUpdateOwner) error
	ClearUpdateOwner(id string) error

	Restart(id string) error
	Build(ids []string, buildParams *body.DeploymentBuild) error
	DoCommand(id string, command string)

	GetCiConfig(id string) (*body.CiConfig, error)

	SetupLogStream(id string, ctx context.Context, handler func(string, string, string), history int) error
	AddLogs(id string, logs ...deploymentModels.Log)

	StartActivity(id string, activity string) error
	CanAddActivity(id, activity string) (bool, string)

	CheckQuota(id string, params *dOpts.QuotaOptions) error
	NameAvailable(name string) (bool, error)
	GetUsage(userID string) (*deploymentModels.Usage, error)

	ValidGitHubToken(token string) (bool, string, error)
	GetGitHubAccessTokenByCode(code string) (string, error)
	GetGitHubRepositories(token string) ([]deploymentModels.GitHubRepository, error)
	ValidGitHubRepository(token string, repositoryID int64) (bool, string, error)

	ValidateHarborToken(secret string) bool

	K8s() *k8s_service.Client
	Harbor() *harbor_service.Client
	GitHub() *github_service.Client
	//K8s() interface {
	//	Create(id string, params *deploymentModels.CreateParams) error
	//	Delete(id string, overwriteUserID ...string) error
	//	Update(id string, params *deploymentModels.UpdateParams) error
	//	EnsureOwner(id string, oldOwnerID string) error
	//	Restart(id string) error
	//	Repair(id string) error
	//	SetupLogStream(ctx context.Context, id string, handler func(string, int, time.Time)) error
	//}
	//Harbor() interface {
	//	Create(id string, params *deploymentModels.CreateParams) error
	//	CreatePlaceholder(id string) error
	//	Delete(id string) error
	//	Update(id string, params *deploymentModels.UpdateParams) error
	//	EnsureOwner(id string, oldOwnerID string) error
	//	Repair(id string) error
	//}
	//GitHub() interface {
	//	Create(id string, params *deploymentModels.CreateParams) error
	//	Delete(id string) error
	//	CreatePlaceholder(id string) error
	//	Validate() (bool, string, error)
	//	GetRepositories() ([]deploymentModels.GitHubRepository, error)
	//	GetRepository() (*deploymentModels.GitHubRepository, error)
	//	GetWebhooks(repository *deploymentModels.GitHubRepository) ([]deploymentModels.GitHubWebhook, error)
	//	GetAccessTokenByCode(code string) (string, error)
	//}
}

type Discovery interface {
	Discover() (*discover.Discover, error)
}

type Events interface {
	Create(id string, params *eventModels.CreateParams) error
}

type Jobs interface {
	Get(id string, opts ...opts.GetOpts) (*jobModels.Job, error)
	List(opts ...opts.ListOpts) ([]jobModels.Job, error)
	Create(id, userID, jobType string, args map[string]interface{}) error
	Update(id string, jobUpdateDTO *body.JobUpdate) (*jobModels.Job, error)
}

type Notifications interface {
	Get(id string, opts ...opts2.GetOpts) (*notificationModels.Notification, error)
	List(opts ...opts2.ListOpts) ([]notificationModels.Notification, error)
	Create(id, userID string, params *notificationModels.CreateParams) (*notificationModels.Notification, error)
	Update(id string, dtoNotificationUpdate *body.NotificationUpdate) (*notificationModels.Notification, error)
	Delete(id string) error
}

type SMs interface {
	Get(id string, opts ...smClient.GetOpts) (*smModels.SM, error)
	GetByUserID(userID string, opts ...smClient.GetOpts) (*smModels.SM, error)
	List(opts ...smClient.ListOpts) ([]smModels.SM, error)
	Create(id, userID string, params *smModels.CreateParams) error
	Delete(id string) error
	Repair(id string) error
	Exists(userID string) (bool, error)

	GetZone() *configModels.DeploymentZone
	GetURL(userID string) *string

	K8s() *sK8sService.Client
}

type Status interface {
	ListWorkerStatus(opts ...opts3.ListWorkerStatusOpts) ([]workerStatusModels.WorkerStatus, error)
}

type Users interface {
	Get(id string, opts ...opts5.GetOpts) (*userModels.User, error)
	List(opts ...opts5.ListOpts) ([]userModels.User, error)
	Create() (*userModels.User, error)
	Update(userID string, dtoUserUpdate *body.UserUpdate) (*userModels.User, error)
	Exists(id string) (bool, error)

	Discover(opts ...opts5.DiscoverOpts) ([]body.UserReadDiscovery, error)
}

type Teams interface {
	Get(id string, opts ...opts4.GetOpts) (*teamModels.Team, error)
	List(opts ...opts4.ListOpts) ([]teamModels.Team, error)
	ListIDs(opts ...opts4.ListOpts) ([]string, error)
	Create(id, ownerID string, dtoCreateTeam *body.TeamCreate) (*teamModels.Team, error)
	Update(id string, dtoUpdateTeam *body.TeamUpdate) (*teamModels.Team, error)
	Delete(id string) error
	Join(id string, dtoTeamJoin *body.TeamJoin) (*teamModels.Team, error)
}

type VMs interface {
	Get(id string, opts ...vmClient.GetOpts) (*vmModels.VM, error)
	List(opts ...vmClient.ListOpts) ([]vmModels.VM, error)
	Create(id, ownerID string, dtoVmCreate *body.VmCreate) error
	Update(id string, dtoVmUpdate *body.VmUpdate) error
	Delete(id string) error
	Repair(id string) error

	GetSnapshot(vmID string, id string, opts ...vmClient.GetSnapshotOpts) (*vmModels.Snapshot, error)
	GetSnapshotByName(vmID string, name string, opts ...vmClient.GetSnapshotOpts) (*vmModels.Snapshot, error)
	ListSnapshots(vmID string, opts ...vmClient.ListSnapshotOpts) ([]vmModels.Snapshot, error)
	CreateSnapshot(id string, params *vmClient.CreateSnapshotOpts) error
	DeleteSnapshot(id, snapshotID string) error
	ApplySnapshot(id, snapshotID string) error

	UpdateOwnerSetup(id string, params *body.VmUpdateOwner) (*string, error)
	UpdateOwner(id string, params *body.VmUpdateOwner) error
	ClearUpdateOwner(id string) error

	GetConnectionString(id string) (*string, error)
	GetExternalPortMapper(vmID string) (map[string]int, error)
	DoCommand(id, command string)

	GetHost(vmID string) (*vmModels.Host, error)
	GetCloudStackHostCapabilities(hostName string, zoneName string) (*vmModels.CloudStackHostCapabilities, error)

	StartActivity(id, activity string) error
	CanAddActivity(vmID, activity string) (bool, string, error)
	NameAvailable(name string) (bool, error)
	HttpProxyNameAvailable(id, name string) (bool, error)

	CheckQuota(id, userID string, quota *roleModels.Quotas, opts ...vmClient.QuotaOpts) error
	GetUsage(userID string) (*vmModels.Usage, error)

	GetGPU(id string, opts ...vmClient.GetGpuOpts) (*gpuModels.GPU, error)
	ListGPUs(opts ...vmClient.ListGpuOpts) ([]gpuModels.GPU, error)
	AttachGPU(id string, gpuIDs []string, leaseDuration float64) error
	DetachGPU(id string) error
	IsGpuPrivileged(id string) (bool, error)
	CheckGpuHardwareAvailable(gpuID string) error
	CheckSuitableHost(id, hostName, zoneName string) error

	CS() *cs_service.Client
	K8s() *vK8sService.Client
}

type Zones interface {
	List(opts ...opts6.ListOpts) ([]zoneModels.Zone, error)
	Get(name, zoneType string) *zoneModels.Zone
}
