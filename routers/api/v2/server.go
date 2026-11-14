package v2

import (
	"github.com/gin-gonic/gin"
	apiV2 "github.com/kthcloud/go-deploy/api/openapi/v2"
)

// Server implements apiV2.ServerInterface by delegating to the v2 Gin handlers.
type Server struct {
	metricsHandler gin.HandlerFunc
	createSM       gin.HandlerFunc
	sseSetup       gin.HandlerFunc
}

// NewServer creates a new v2 ServerInterface implementation.
func NewServer(metricsHandler, createSM, sseSetup gin.HandlerFunc) *Server {
	return &Server{
		metricsHandler: metricsHandler,
		createSM:       createSM,
		sseSetup:       sseSetup,
	}
}

var _ apiV2.ServerInterface = (*Server)(nil)

func (s *Server) GetDeployments(c *gin.Context, _ apiV2.GetDeploymentsParams) { ListDeployments(c) }
func (s *Server) PostDeployments(c *gin.Context) {
	s.createSM(c)
	if !c.IsAborted() {
		CreateDeployment(c)
	}
}
func (s *Server) DeleteDeploymentsByDeploymentId(c *gin.Context, _ string) { DeleteDeployment(c) }
func (s *Server) GetDeploymentsByDeploymentId(c *gin.Context, _ string)    { GetDeployment(c) }
func (s *Server) PostDeploymentsByDeploymentId(c *gin.Context, _ string)   { UpdateDeployment(c) }
func (s *Server) GetDeploymentsByDeploymentIdCiConfig(c *gin.Context, _ string) {
	GetCiConfig(c)
}
func (s *Server) PostDeploymentsByDeploymentIdCommand(c *gin.Context, _ string) {
	DoDeploymentCommand(c)
}
func (s *Server) GetDeploymentsByDeploymentIdLogsSse(c *gin.Context, _ string) {
	s.sseSetup(c)
	if !c.IsAborted() {
		GetLogs(c)
	}
}

func (s *Server) GetDiscover(c *gin.Context) { Discover(c) }

func (s *Server) GetGpuClaims(c *gin.Context, _ apiV2.GetGpuClaimsParams) { ListGpuClaims(c) }
func (s *Server) PostGpuClaims(c *gin.Context)                            { CreateGpuClaim(c) }
func (s *Server) DeleteGpuClaimsByGpuClaimId(c *gin.Context, _ string)    { DeleteGpuClaim(c) }
func (s *Server) GetGpuClaimsByGpuClaimId(c *gin.Context, _ string)       { GetGpuClaim(c) }
func (s *Server) GetGpuGroups(c *gin.Context, _ apiV2.GetGpuGroupsParams) { ListGpuGroups(c) }
func (s *Server) GetGpuGroupsByGpuGroupId(c *gin.Context, _ string)       { GetGpuGroup(c) }
func (s *Server) GetGpuLeases(c *gin.Context, _ apiV2.GetGpuLeasesParams) { ListGpuLeases(c) }
func (s *Server) PostGpuLeases(c *gin.Context)                            { CreateGpuLease(c) }
func (s *Server) DeleteGpuLeasesByGpuLeaseId(c *gin.Context, _ string)    { DeleteGpuLease(c) }
func (s *Server) GetGpuLeasesByGpuLeaseId(c *gin.Context, _ string)       { GetGpuLease(c) }
func (s *Server) PostGpuLeasesByGpuLeaseId(c *gin.Context, _ string)      { UpdateGpuLease(c) }
func (s *Server) PostHooksDeploymentsHarbor(c *gin.Context, _ apiV2.PostHooksDeploymentsHarborParams) {
	HandleHarborHook(c)
}
func (s *Server) GetHosts(c *gin.Context)                       { ListHosts(c) }
func (s *Server) GetHostsVerbose(c *gin.Context)                { VerboseListHosts(c) }
func (s *Server) GetJobs(c *gin.Context, _ apiV2.GetJobsParams) { ListJobs(c) }
func (s *Server) GetJobsByJobId(c *gin.Context, _ string)       { GetJob(c) }
func (s *Server) PostJobsByJobId(c *gin.Context, _ string)      { UpdateJob(c) }
func (s *Server) GetMetrics(c *gin.Context)                     { s.metricsHandler(c) }
func (s *Server) GetNotifications(c *gin.Context, _ apiV2.GetNotificationsParams) {
	ListNotifications(c)
}
func (s *Server) DeleteNotificationsByNotificationId(c *gin.Context, _ string) {
	DeleteNotification(c)
}
func (s *Server) GetNotificationsByNotificationId(c *gin.Context, _ string) {
	GetNotification(c)
}
func (s *Server) PostNotificationsByNotificationId(c *gin.Context, _ string) {
	UpdateNotification(c)
}
func (s *Server) PostRegister(c *gin.Context) { Register(c) }
func (s *Server) GetResourceMigrations(c *gin.Context, _ apiV2.GetResourceMigrationsParams) {
	ListResourceMigrations(c)
}
func (s *Server) PostResourceMigrations(c *gin.Context) { CreateResourceMigration(c) }
func (s *Server) DeleteResourceMigrationsByResourceMigrationId(c *gin.Context, _ string) {
	DeleteResourceMigration(c)
}
func (s *Server) GetResourceMigrationsByResourceMigrationId(c *gin.Context, _ string) {
	GetResourceMigration(c)
}
func (s *Server) PostResourceMigrationsByResourceMigrationId(c *gin.Context, _ string) {
	UpdateResourceMigration(c)
}
func (s *Server) GetSnapshots(c *gin.Context, _ apiV2.GetSnapshotsParams) { ListSnapshots(c) }
func (s *Server) PostSnapshots(c *gin.Context)                            { CreateSnapshot(c) }
func (s *Server) DeleteSnapshotsBySnapshotId(c *gin.Context, _ string)    { DeleteSnapshot(c) }
func (s *Server) GetSnapshotsBySnapshotId(c *gin.Context, _ string)       { GetSnapshot(c) }
func (s *Server) GetStorageManagers(c *gin.Context, _ apiV2.GetStorageManagersParams) {
	ListSMs(c)
}
func (s *Server) DeleteStorageManagersByStorageManagerId(c *gin.Context, _ string) {
	DeleteSM(c)
}
func (s *Server) GetStorageManagersByStorageManagerId(c *gin.Context, _ string) {
	GetSM(c)
}
func (s *Server) GetSystemCapacities(c *gin.Context, _ apiV2.GetSystemCapacitiesParams) {
	ListSystemCapacities(c)
}
func (s *Server) GetSystemStats(c *gin.Context, _ apiV2.GetSystemStatsParams) {
	ListSystemStats(c)
}
func (s *Server) GetSystemStatus(c *gin.Context, _ apiV2.GetSystemStatusParams) {
	ListSystemStatus(c)
}
func (s *Server) GetTeams(c *gin.Context, _ apiV2.GetTeamsParams) { ListTeams(c) }
func (s *Server) PostTeams(c *gin.Context)                        { CreateTeam(c) }
func (s *Server) DeleteTeamsByTeamId(c *gin.Context, _ string)    { DeleteTeam(c) }
func (s *Server) GetTeamsByTeamId(c *gin.Context, _ string)       { GetTeam(c) }
func (s *Server) PostTeamsByTeamId(c *gin.Context, _ string)      { UpdateTeam(c) }
func (s *Server) GetUsers(c *gin.Context, _ apiV2.GetUsersParams) { ListUsers(c) }
func (s *Server) PostUsers(c *gin.Context)                        { UpdateUser(c) }
func (s *Server) GetUsersByUserId(c *gin.Context, _ string, _ apiV2.GetUsersByUserIdParams) {
	GetUser(c)
}
func (s *Server) PostUsersByUserId(c *gin.Context, _ string)        { UpdateUser(c) }
func (s *Server) PostUsersByUserIdApiKeys(c *gin.Context, _ string) { CreateApiKey(c) }
func (s *Server) PostVmActions(c *gin.Context)                      { CreateVmAction(c) }
func (s *Server) GetVms(c *gin.Context, _ apiV2.GetVmsParams)       { ListVMs(c) }
func (s *Server) PostVms(c *gin.Context)                            { CreateVM(c) }
func (s *Server) DeleteVmsByVmId(c *gin.Context, _ string)          { DeleteVM(c) }
func (s *Server) GetVmsByVmId(c *gin.Context, _ string)             { GetVM(c) }
func (s *Server) PostVmsByVmId(c *gin.Context, _ string)            { UpdateVM(c) }
func (s *Server) GetWorkerStatus(c *gin.Context)                    { ListWorkerStatus(c) }
func (s *Server) GetZones(c *gin.Context)                           { ListZones(c) }
