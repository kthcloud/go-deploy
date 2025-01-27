// Code generated by tygo. DO NOT EDIT.

//////////
// source: common.go

export interface Pagination {
  Page: number /* int */;
  PageSize: number /* int */;
}
export interface SortBy {
  Field: string;
  Order: number /* int */;
}

//////////
// source: deployment.go

export interface Env {
  key: string;
  val: string;
}
export interface DeploymentGet {
  /**
   * MigrationCode is used when fetching a deployment that is being migrated.
   * The token should only be known by the user receiving the deployment.
   */
  migrationCode?: string;
}
export interface DeploymentList {
  Pagination?: Pagination;
  All: boolean;
  UserID?: string;
}
export interface DeploymentUpdate {
  envs: { [key: string]: string}[];
}

//////////
// source: gpu_group.go

export interface GpuGroupList {
  Pagination?: Pagination;
}

//////////
// source: gpu_lease.go

export interface GpuLeaseList {
  Pagination?: Pagination;
  All: boolean;
  VmIDs: string[];
}
export interface GpuLeaseCreate {
}

//////////
// source: job.go

export interface JobList {
  Pagination?: Pagination;
  SortBy?: SortBy;
  All: boolean;
  Status: string[];
  ExcludeStatus: string[];
  Types: string[];
  ExcludeTypes: string[];
  UserID?: string;
}

//////////
// source: notification.go

export interface NotificationList {
  Pagination?: Pagination;
  All: boolean;
  UserID?: string;
}

//////////
// source: resource_migration.go

export interface ResourceMigrationList {
  Pagination?: Pagination;
}

//////////
// source: sm.go

export interface SmList {
  Pagination?: Pagination;
  All: boolean;
}

//////////
// source: snapshot.go

export interface VmSnapshotList {
  Pagination?: Pagination;
}

//////////
// source: status.go

export interface StatusList {
}

//////////
// source: team.go

export interface TeamList {
  Pagination?: Pagination;
  UserID?: string;
  All: boolean;
}

//////////
// source: timestamp_request.go

export interface TimestampRequest {
  N: number /* int */;
}

//////////
// source: user.go

export interface UserGet {
  Discover: boolean;
}
export interface UserList {
  Pagination?: Pagination;
  All: boolean;
  Search?: string;
  Discover: boolean;
}

//////////
// source: vm.go

export interface VmGet {
  /**
   * MigrationCode is used when fetching a deployment that is being migrated.
   * The token should only be known by the user receiving the deployment.
   */
  migrationCode?: string;
}
export interface VmList {
  Pagination?: Pagination;
  All: boolean;
  UserID?: string;
}

//////////
// source: vm_action.go

export interface VmActionCreate {
  VmID: string;
}

//////////
// source: zone.go

export interface ZoneList {
}
