// Code generated by tygo. DO NOT EDIT.

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
  url?: string;
  envs: Env[];
  volumes: Volume[];
  initCommands: string[];
  args: string[];
  private: boolean;
  internalPort: number /* int */;
  image?: string;
  healthCheckPath?: string;
  replicas: number /* int */;
  customDomain?: string;
  customDomainUrl?: string;
  customDomainStatus?: string;
  customDomainSecret?: string;
  status: string;
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
  image?: string;
  private: boolean;
  envs: Env[];
  volumes: Volume[];
  initCommands: string[];
  args: string[];
  healthCheckPath?: string;
  customDomain?: string;
  replicas?: number /* int */;
  zone?: string;
}
export interface DeploymentUpdate {
  /**
   * update
   */
  name?: string;
  private?: boolean;
  envs?: Env[];
  volumes?: Volume[];
  initCommands?: string[];
  args?: string[];
  customDomain?: string;
  image?: string;
  healthCheckPath?: string;
  replicas?: number /* int */;
  /**
   * update owner
   */
  ownerId?: string;
  transferCode?: string;
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
export interface DeploymentUpdateOwner {
  newOwnerId: string;
  oldOwnerId: string;
  transferCode?: string;
}
export interface DeploymentBuild {
  Name: string;
  Tag: string;
  Branch: string;
  ImportURL: string;
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
// source: gpu.go

export interface GpuAttached {
  id: string;
  jobId: string;
}
export interface GpuDetached {
  id: string;
  jobId: string;
}
export interface Lease {
  vmId?: string;
  user?: string;
  end: string;
  expired: boolean;
}
export interface GpuRead {
  id: string;
  name: string;
  zone: string;
  lease?: Lease;
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
  completedAt?: string;
}
export interface NotificationUpdate {
  read: boolean;
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
// source: status.go

export interface WorkerStatusRead {
  name: string;
  status: string;
  reported_at: string;
}

//////////
// source: team.go

export interface TeamMember {
  id: string;
  username: string;
  email: string;
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

export interface PublicKey {
  name: string;
  key: string;
}
export interface Quota {
  deployments: number /* int */;
  cpuCores: number /* int */;
  ram: number /* int */;
  diskSize: number /* int */;
  snapshots: number /* int */;
  gpuLeaseDuration: number /* float64 */; // in hours
}
export interface Usage {
  deployments: number /* int */;
  cpuCores: number /* int */;
  ram: number /* int */;
  diskSize: number /* int */;
  snapshots: number /* int */;
}
export interface SmallUserRead {
  id: string;
  username: string;
  firstName: string;
  lastName: string;
  email: string;
}
export interface UserRead {
  id: string;
  username: string;
  firstName: string;
  lastName: string;
  email: string;
  publicKeys: PublicKey[];
  onboarded: boolean;
  role: Role;
  admin: boolean;
  quota: Quota;
  usage: Usage;
  storageUrl?: string;
}
export interface UserReadDiscovery {
  id: string;
  username: string;
  firstName: string;
  lastName: string;
  email: string;
}
export interface UserUpdate {
  publicKeys?: PublicKey[];
  onboarded?: boolean;
}

//////////
// source: user_data.go

export interface UserDataRead {
  id: string;
  userId: string;
  data: string;
}
export interface UserDataCreate {
  id: string;
  data: string;
}
export interface UserDataUpdate {
  data: string;
}

//////////
// source: vm.go

export interface VmRead {
  id: string;
  name: string;
  ownerId: string;
  zone: string;
  host?: string;
  createdAt: string;
  updatedAt?: string;
  repairedAt?: string;
  specs?: Specs;
  ports: PortRead[];
  gpu_repo?: VmGpuLease;
  sshPublicKey: string;
  teams: string[];
  status: string;
  connectionString?: string;
}
export interface VmCreate {
  name: string;
  sshPublicKey: string;
  ports: PortCreate[];
  cpuCores: number /* int */;
  ram: number /* int */;
  diskSize: number /* int */;
  zone?: string;
}
export interface VmUpdate {
  name?: string;
  snapshotId?: string;
  ports?: PortUpdate[];
  cpuCores?: number /* int */;
  ram?: number /* int */;
  gpuId?: string;
  noLeaseEnd?: boolean;
  ownerId?: string;
  transferCode?: string;
}
export interface VmUpdateOwner {
  newOwnerId: string;
  oldOwnerId: string;
  transferCode?: string;
}
export interface VmGpuLease {
  id: string;
  name: string;
  leaseEnd: string;
  expired: boolean;
}
export interface Specs {
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
// source: vm_command.go

export interface VmCommand {
  command: string;
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
export interface HttpProxyRead {
  name: string;
  url?: string;
  customDomain?: string;
  customDomainUrl?: string;
  customDomainStatus?: string;
  customDomainSecret?: string;
}
export interface HttpProxyCreate {
  name: string;
  customDomain?: string;
}
export interface HttpProxyUpdate {
  name?: string;
  customDomain?: string;
}

//////////
// source: vm_snapshot.go

export interface VmSnapshotRead {
  id: string;
  vmId: string;
  displayName: string;
  parentName?: string;
  created: string;
  state: string;
  current: boolean;
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
export interface GitHubWebhookPing {
  hook: {
    id: number /* int64 */;
    type: string;
    events: string[];
    Config: {
      url: string;
      content_type: string;
      token: string;
    };
  };
  repository: {
    id: number /* int64 */;
    name: string;
    Owner: {
      id: number /* int64 */;
      login: string;
    };
  };
}
export interface GithubWebhookPayloadPush {
  ref: string;
  repository: {
    id: number /* int64 */;
    name: string;
    Owner: {
      id: number /* int64 */;
      login: string;
    };
    clone_url: string;
    default_branch: string;
  };
}
export interface GitHubWebhookPush {
  ID: number /* int64 */;
  Event: string;
  Signature: string;
  Payload: GithubWebhookPayloadPush;
}

//////////
// source: zone.go

export interface ZoneRead {
  name: string;
  description: string;
  type: string;
  interface?: string;
}
