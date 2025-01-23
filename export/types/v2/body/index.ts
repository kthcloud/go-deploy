// Code generated by tygo. DO NOT EDIT.

//////////
// source: api_key.go

export interface ApiKeyCreate {
  name: string;
  expiresAt: string;
}
export interface ApiKeyCreated {
  name: string;
  key: string;
  createdAt: string;
  expiresAt: string;
}

//////////
// source: cluster.go

export interface ClusterRegisterParams {
}

//////////
// source: deployment.go

export interface DeploymentRead {
  id: string;
  name: string;
  type: string;
  ownerId: string;
  zone: string;
  createdAt: string;
  updatedAt?: string;
  repairedAt?: string;
  restartedAt?: string;
  accessedAt: string;
  url?: string;
  specs: DeploymentSpecs;
  envs: Env[];
  volumes: Volume[];
  initCommands: string[];
  args: string[];
  internalPort: number /* int */;
  image?: string;
  healthCheckPath?: string;
  customDomain?: CustomDomainRead;
  visibility: string;
  neverStale: boolean;
  /**
   * Deprecated: Use Visibility instead.
   */
  private: boolean;
  status: string;
  error?: string;
  replicaStatus?: ReplicaStatus;
  pingResult?: number /* int */;
  /**
   * Integrations are currently not used, but could be used if we wanted to add a list of integrations to the deployment
   * For example GitHub
   */
  integrations: string[];
  teams: string[];
  storageUrl?: string;
}
export interface DeploymentCreate {
  name: string;
  cpuCores?: number /* float64 */;
  ram?: number /* float64 */;
  replicas?: number /* int */;
  envs: Env[];
  volumes: Volume[];
  initCommands: string[];
  args: string[];
  visibility: string;
  /**
   * Boolean to make deployment never get disabled, despite being stale
   */
  neverStale: boolean;
  /**
   * Deprecated: Use Visibility instead.
   */
  private: boolean;
  image?: string;
  healthCheckPath?: string;
  /**
   * CustomDomain is the domain that the deployment will be available on.
   * The max length is set to 243 to allow for a subdomain when confirming the domain.
   */
  customDomain?: string;
  /**
   * Zone is the zone that the deployment will be created in.
   * If the zone is not set, the deployment will be created in the default zone.
   */
  zone?: string;
}
export interface DeploymentUpdate {
  name?: string;
  cpuCores?: number /* float64 */;
  ram?: number /* float64 */;
  replicas?: number /* int */;
  envs?: Env[];
  volumes?: Volume[];
  initCommands?: string[];
  args?: string[];
  visibility?: string;
  neverStale?: boolean;
  /**
   * Deprecated: Use Visibility instead.
   */
  private?: boolean;
  image?: string;
  healthCheckPath?: string;
  /**
   * CustomDomain is the domain that the deployment will be available on.
   * The max length is set to 243 to allow for a subdomain when confirming the domain.
   */
  customDomain?: string;
}
export interface Env {
  name: string;
  value: string;
}
export interface Volume {
  name: string;
  appPath: string;
  serverPath: string;
}
export interface DeploymentBuild {
  Name: string;
  Tag: string;
  Branch: string;
  ImportURL: string;
}
export interface ReplicaStatus {
  /**
   * DesiredReplicas is the number of replicas that the deployment should have.
   */
  desiredReplicas: number /* int */;
  /**
   * ReadyReplicas is the number of replicas that are ready.
   */
  readyReplicas: number /* int */;
  /**
   * AvailableReplicas is the number of replicas that are available.
   */
  availableReplicas: number /* int */;
  /**
   * UnavailableReplicas is the number of replicas that are unavailable.
   */
  unavailableReplicas: number /* int */;
}
export interface DeploymentCreated {
  id: string;
  jobId: string;
}
export interface DeploymentDeleted {
  id: string;
  jobId: string;
}
export interface DeploymentUpdated {
  id: string;
  jobId?: string;
}
export interface DeploymentSpecs {
  cpuCores: number /* float64 */;
  ram: number /* float64 */;
  replicas: number /* int */;
}
export interface CiConfig {
  config: string;
}
export interface DeploymentCommand {
  command: string;
}
export interface LogMessage {
  source: string;
  prefix: string;
  line: string;
  createdAt: string;
}

//////////
// source: discover.go

export interface DiscoverRead {
  version: string;
  roles: Role[];
}

//////////
// source: error.go

export interface BindingError {
  validationErrors: { [key: string]: string[]};
}

//////////
// source: gpu_group.go

export interface GpuGroupRead {
  id: string;
  name: string;
  displayName: string;
  zone: string;
  vendor: string;
  total: number /* int */;
  available: number /* int */;
}

//////////
// source: gpu_lease.go

export interface GpuLeaseGpuGroup {
  id: string;
  name: string;
  displayName: string;
}
export interface GpuLeaseRead {
  id: string;
  gpuGroupId: string;
  active: boolean;
  userId: string;
  /**
   * VmID is set when the lease is attached to a VM.
   */
  vmId?: string;
  queuePosition: number /* int */;
  leaseDuration: number /* float64 */;
  /**
   * ActivatedAt specifies the time when the lease was activated. This is the time the user first attached the GPU
   * or 1 day after the lease was created if the user did not attach the GPU.
   */
  activatedAt?: string;
  /**
   * AssignedAt specifies the time when the lease was assigned to the user.
   */
  assignedAt?: string;
  createdAt: string;
  /**
   * ExpiresAt specifies the time when the lease will expire.
   * This is only present if the lease is active.
   */
  expiresAt?: string;
  expiredAt?: string;
}
export interface GpuLeaseCreate {
  /**
   * GpuGroupID is used to specify the GPU to lease.
   * As such, the lease does not specify which specific GPU to lease, but rather the type of GPU to lease.
   */
  gpuGroupId: string;
  /**
   * LeaseForever is used to specify whether the lease should be created forever.
   */
  leaseForever: boolean;
}
export interface GpuLeaseUpdate {
  /**
   * VmID is used to specify the VM to attach the lease to.
   * - If specified, the lease will be attached to the VM.
   * - If the lease is already attached to a VM, it will be detached from the current VM and attached to the new VM.
   * - If the lease is not active, specifying a VM will activate the lease.
   * - If the lease is not assigned, an error will be returned.
   */
  vmId?: string;
}
export interface GpuLeaseCreated {
  id: string;
  jobId: string;
}
export interface GpuLeaseUpdated {
  id: string;
  jobId: string;
}
export interface GpuLeaseDeleted {
  id: string;
  jobId: string;
}

//////////
// source: host.go

export interface HostRead extends HostBase {
}
export interface HostBase {
  name: string;
  displayName: string;
  /**
   * Zone is the name of the zone where the host is located.
   */
  zone: string;
}
export interface HostRegisterParams {
  /**
   * Name is the host name of the node
   */
  name: string;
  /**
   * DisplayName is the human-readable name of the node
   * This is optional, and is set to Name if not provided
   */
  displayName: string;
  ip: string;
  /**
   * Port is the port the node is listening on for API requests
   */
  port: number /* int */;
  zone: string;
  /**
   * Token is the discovery token validated against the config
   */
  token: string;
  /**
   * Enabled is the flag to enable or disable the node
   */
  enabled: boolean;
  /**
   * Schedulable is the flag to enable or disable scheduling on the node
   */
  schedulable: boolean;
}

//////////
// source: job.go

export interface JobRead {
  id: string;
  userId: string;
  type: string;
  status: string;
  lastError?: string;
  createdAt: string;
  lastRunAt?: string;
  finishedAt?: string;
  runAfter?: string;
}
export interface JobUpdate {
  status?: string;
}

//////////
// source: notification.go

export interface NotificationRead {
  id: string;
  userId: string;
  type: string;
  content: { [key: string]: any};
  createdAt: string;
  readAt?: string;
  toastedAt?: string;
  completedAt?: string;
}
export interface NotificationUpdate {
  read: boolean;
  toasted: boolean;
}

//////////
// source: resource_migration.go

export interface ResourceMigrationRead {
  id: string;
  /**
   * ResourceID is the ID of the resource that is being migrated.
   * This can be a VM ID, deployment ID, etc. depending on the type of the migration.
   */
  resourceId: string;
  /**
   * UserID is the ID of the user who initiated the migration.
   */
  userId: string;
  /**
   * Type is the type of the resource migration.
   * Possible values:
   * - updateOwner
   */
  type: string;
  /**
   * ResourceType is the type of the resource that is being migrated.
   * Possible values:
   * - vm
   * - deployment
   */
  resourceType: string;
  /**
   * Status is the status of the resource migration.
   * When this field is set to 'accepted', the migration will take place and then automatically be deleted.
   */
  status: string;
  /**
   * UpdateOwner is the set of parameters that are required for the updateOwner migration type.
   * It is empty if the migration type is not updateOwner.
   */
  updateOwner?: {
    ownerId: string;
  };
  createdAt: string;
  deletedAt?: string;
}
export interface ResourceMigrationCreate {
  /**
   * Type is the type of the resource migration.
   * Possible values:
   * - updateOwner
   */
  type: string;
  /**
   * ResourceID is the ID of the resource that is being migrated.
   * This can be a VM ID, deployment ID, etc. depending on the type of the migration.
   */
  resourceId: string;
  /**
   * Status is the status of the resource migration.
   * It is used by privileged admins to directly accept or reject a migration.
   * The field is ignored by non-admins.
   * Possible values:
   * - accepted
   * - pending
   */
  status?: string;
  /**
   * UpdateOwner is the set of parameters that are required for the updateOwner migration type.
   * It is ignored if the migration type is not updateOwner.
   */
  updateOwner?: {
    ownerId: string;
  };
}
export interface ResourceMigrationUpdate {
  /**
   * Status is the status of the resource migration.
   * It is used to accept a migration by setting the status to 'accepted'.
   * If the acceptor is not an admin, a Code must be provided.
   * Possible values:
   * - accepted
   * - pending
   */
  status: string;
  /**
   * Code is a token required when accepting a migration if the acceptor is not an admin.
   * It is sent to the acceptor using the notification API
   */
  code?: string;
}
export interface ResourceMigrationCreated extends ResourceMigrationRead {
  /**
   * JobID is the ID of the job that was created for the resource migration.
   * It will only be set if the migration was created with status 'accepted'.
   */
  jobId?: string;
}
export interface ResourceMigrationUpdated extends ResourceMigrationRead {
  /**
   * JobID is the ID of the job that was created for the resource migration.
   * It will only be set if the migration was updated with status 'accepted'.
   */
  jobId?: string;
}

//////////
// source: role.go

export interface Role {
  name: string;
  description: string;
  permissions: string[];
  quota?: Quota;
}

//////////
// source: sm.go

export interface SmDeleted {
  id: string;
  jobId: string;
}
export interface SmRead {
  id: string;
  ownerId: string;
  createdAt: string;
  zone: string;
  url?: string;
}

//////////
// source: snapshot.go

export interface VmSnapshotRead {
  id: string;
  name: string;
  status: string;
  created: string;
}
export interface VmSnapshotCreate {
  name: string;
}
export interface VmSnapshotCreated {
  id: string;
  jobId: string;
}
export interface VmSnapshotDeleted {
  id: string;
  jobId: string;
}

//////////
// source: status.go

export interface WorkerStatusRead {
  name: string;
  status: string;
  reportedAt: string;
}

//////////
// source: system_capacities.go

export interface TimestampedSystemCapacities {
  capacities: SystemCapacities;
  timestamp: string;
}
export interface SystemCapacities {
  /**
   * Total
   */
  cpuCore: CpuCoreCapacities;
  ram: RamCapacities;
  gpu: GpuCapacities;
  /**
   * Per Host
   */
  hosts: HostCapacities[];
  /**
   * Per Cluster
   */
  clusters: ClusterCapacities[];
}
export interface ClusterCapacities {
  cluster: string;
  cpuCore: CpuCoreCapacities;
  ram: RamCapacities;
  gpu: GpuCapacities;
}
export interface HostCapacities extends HostBase {
  cpuCore: CpuCoreCapacities;
  ram: RamCapacities;
  gpu: GpuCapacities;
}
export interface CpuCoreCapacities {
  total: number /* int */;
}
export interface RamCapacities {
  total: number /* int */;
}
export interface GpuCapacities {
  total: number /* int */;
}

//////////
// source: system_gpu_info.go

export interface SystemGpuInfo {
  hosts: HostGpuInfo[];
}
export interface TimestampedSystemGpuInfo {
  gpuInfo: SystemGpuInfo;
  timestamp: string;
}
export interface HostGpuInfo extends HostBase {
  gpus: GpuInfo[];
}
export interface GpuInfo {
  name: string;
  slot: string;
  vendor: string;
  vendorId: string;
  bus: string;
  deviceId: string;
  passthrough: boolean;
}

//////////
// source: system_stats.go

export interface SystemStats {
  k8s: K8sStats;
}
export interface TimestampedSystemStats {
  stats: SystemStats;
  timestamp: string;
}
export interface K8sStats {
  podCount: number /* int */;
  clusters: ClusterStats[];
}
export interface ClusterStats {
  cluster: string;
  podCount: number /* int */;
}

//////////
// source: system_status.go

export interface SystemStatus {
  hosts: HostStatus[];
}
export interface TimestampedSystemStatus {
  status: SystemStatus;
  timestamp: string;
}
export interface HostStatus extends HostBase {
  cpu: CpuStatus;
  ram: RamStatus;
  gpu?: GpuStatus;
}
export interface CpuStatus {
  temp: CpuStatusTemp;
  load: CpuStatusLoad;
}
export interface CpuStatusTemp {
  main: number /* float64 */;
  cores: number /* int */[];
  max: number /* float64 */;
}
export interface CpuStatusLoad {
  main: number /* float64 */;
  cores: number /* int */[];
  max: number /* float64 */;
}
export interface RamStatus {
  load: RamStatusLoad;
}
export interface RamStatusLoad {
  main: number /* float64 */;
}
export interface GpuStatus {
  temp: GpuStatusTemp[];
}
export interface GpuStatusTemp {
  main: number /* float64 */;
}

//////////
// source: team.go

export interface TeamMember extends UserReadDiscovery {
  teamRole: string;
  memberStatus: string;
  joinedAt?: string;
  addedAt?: string;
}
export interface TeamResource {
  id: string;
  name: string;
  type: string;
}
export interface TeamMemberCreate {
  id: string;
  teamRole: string; // default to MemberRoleAdmin right now
}
export interface TeamMemberUpdate {
  id: string;
  teamRole: string; // default to MemberRoleAdmin right now
}
export interface TeamCreate {
  name: string;
  description: string;
  resources: string[];
  members: TeamMemberCreate[];
}
export interface TeamJoin {
  invitationCode: string;
}
export interface TeamUpdate {
  name?: string;
  description?: string;
  resources?: string[];
  members?: TeamMemberUpdate[];
}
export interface TeamRead {
  id: string;
  name: string;
  ownerId: string;
  description?: string;
  resources: TeamResource[];
  members: TeamMember[];
  createdAt: string;
  updatedAt?: string;
}

//////////
// source: user.go

export interface UserRead {
  id: string;
  username: string;
  firstName: string;
  lastName: string;
  email: string;
  publicKeys: PublicKey[];
  apiKeys: ApiKey[];
  userData: UserData[];
  role: Role;
  admin: boolean;
  quota: Quota;
  usage: Usage;
  storageUrl?: string;
  gravatarUrl?: string;
}
export interface UserReadDiscovery {
  id: string;
  username: string;
  firstName: string;
  lastName: string;
  email: string;
  gravatarUrl?: string;
}
export interface UserUpdate {
  publicKeys?: PublicKey[];
  /**
   * ApiKeys specifies the API keys that should remain. If an API key is not in this list, it will be deleted.
   * However, API keys cannot be created, use /apiKeys endpoint to create new API keys.
   */
  apiKeys?: ApiKey[];
  userData?: UserData[];
}
export interface UserData {
  key: string;
  value: string;
}
export interface PublicKey {
  name: string;
  key: string;
}
export interface ApiKey {
  name: string;
  createdAt: string;
  expiresAt: string;
}
export interface Quota {
  cpuCores: number /* float64 */;
  ram: number /* float64 */;
  diskSize: number /* float64 */;
  snapshots: number /* int */;
  gpuLeaseDuration: number /* float64 */; // in hours
}
export interface Usage {
  cpuCores: number /* float64 */;
  ram: number /* float64 */;
  diskSize: number /* int */;
}

//////////
// source: vm.go

export interface VmRead {
  id: string;
  name: string;
  internalName?: string;
  ownerId: string;
  zone: string;
  host?: string;
  createdAt: string;
  updatedAt?: string;
  repairedAt?: string;
  accessedAt: string;
  neverStale: boolean;
  specs: VmSpecs;
  ports: PortRead[];
  gpu?: VmGpuLease;
  sshPublicKey: string;
  teams: string[];
  status: string;
  sshConnectionString?: string;
}
export interface VmCreate {
  name: string;
  sshPublicKey: string;
  ports: PortCreate[];
  cpuCores: number /* int */;
  ram: number /* int */;
  diskSize: number /* int */;
  zone?: string;
  neverStale: boolean;
}
export interface VmUpdate {
  name?: string;
  ports?: PortUpdate[];
  cpuCores?: number /* int */;
  ram?: number /* int */;
  neverStale?: boolean;
}
export interface VmUpdateOwner {
  newOwnerId: string;
  oldOwnerId: string;
}
export interface VmGpuLease {
  id: string;
  gpuGroupId: string;
  leaseDuration: number /* float64 */;
  /**
   * ActivatedAt specifies the time when the lease was activated. This is the time the user first attached the GPU
   * or 1 day after the lease was created if the user did not attach the GPU.
   */
  activatedAt?: string;
  /**
   * AssignedAt specifies the time when the lease was assigned to the user.
   */
  assignedAt?: string;
  createdAt: string;
  /**
   * ExpiresAt specifies the time when the lease will expire.
   * This is only present if the lease is active.
   */
  expiresAt?: string;
  /**
   * ExpiredAt specifies the time when the lease expired.
   * This is only present if the lease is expired.
   */
  expiredAt?: string;
}
export interface VmSpecs {
  cpuCores?: number /* int */;
  ram?: number /* int */;
  diskSize?: number /* int */;
}
export interface VmCreated {
  id: string;
  jobId: string;
}
export interface VmDeleted {
  id: string;
  jobId: string;
}
export interface VmUpdated {
  id: string;
  jobId?: string;
}

//////////
// source: vm_action.go

export interface VmActionCreate {
  action: string;
}
export interface VmActionCreated {
  id: string;
  jobId: string;
}

//////////
// source: vm_port.go

export interface PortRead {
  name?: string;
  port?: number /* int */;
  externalPort?: number /* int */;
  protocol?: string;
  httpProxy?: HttpProxyRead;
}
export interface PortCreate {
  name: string;
  port: number /* int */;
  protocol: string;
  httpProxy?: HttpProxyCreate;
}
export interface PortUpdate {
  name?: string;
  port?: number /* int */;
  protocol?: string;
  httpProxy?: HttpProxyUpdate;
}
export interface CustomDomainRead {
  domain: string;
  url: string;
  status: string;
  secret: string;
}
export interface HttpProxyRead {
  name: string;
  url?: string;
  customDomain?: CustomDomainRead;
}
export interface HttpProxyCreate {
  name: string;
  /**
   * CustomDomain is the domain that the deployment will be available on.
   * The max length is set to 243 to allow for a subdomain when confirming the domain.
   */
  customDomain?: string;
}
export interface HttpProxyUpdate {
  name?: string;
  /**
   * CustomDomain is the domain that the deployment will be available on.
   * The max length is set to 243 to allow for a subdomain when confirming the domain.
   */
  customDomain?: string;
}

//////////
// source: webhook.go

export interface HarborWebhook {
  type: string;
  occur_at: number /* int */;
  operator: string;
  event_data: {
    resources: {
      digest: string;
      tag: string;
      resource_url: string;
    }[];
    repository: {
      date_created: number /* int */;
      name: string;
      namespace: string;
      repo_full_name: string;
      repo_type: string;
    };
  };
}

//////////
// source: zone.go

export interface ZoneEndpoints {
  deployment?: string;
  storage?: string;
  vm?: string;
  vmApp?: string;
}
export interface ZoneRead {
  name: string;
  description: string;
  capabilities: string[];
  endpoints: ZoneEndpoints;
  legacy: boolean;
  enabled: boolean;
}
