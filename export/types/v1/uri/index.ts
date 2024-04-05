// Code generated by tygo. DO NOT EDIT.

//////////
// source: deployment.go

export interface DeploymentGet {
  DeploymentID: string;
}
export interface DeploymentDelete {
  DeploymentID: string;
}
export interface DeploymentUpdate {
  DeploymentID: string;
}
export interface CiConfigGet {
  DeploymentID: string;
}
export interface DeploymentCommand {
  DeploymentID: string;
}
export interface LogsGet {
  DeploymentID: string;
}
export interface BuildGet {
  DeploymentID: string;
}

//////////
// source: job.go

export interface JobGet {
  JobID: string;
}
export interface JobUpdate {
  JobID: string;
}

//////////
// source: notification.go

export interface NotificationGet {
  NotificationID: string;
}
export interface NotificationUpdate {
  NotificationID: string;
}
export interface NotificationDelete {
  NotificationID: string;
}

//////////
// source: sm.go

export interface SmGet {
  SmID: string;
}
export interface SmDelete {
  SmID: string;
}

//////////
// source: team.go

export interface TeamGet {
  TeamID: string;
}
export interface TeamUpdate {
  TeamID: string;
}

//////////
// source: user.go

export interface UserGet {
  UserID: string;
}
export interface UserUpdate {
  UserID: string;
}

//////////
// source: user_data.go

export interface UserDataGet {
  ID: string;
}
export interface UserDataUpdate {
  ID: string;
}
export interface UserDataDelete {
  ID: string;
}

//////////
// source: vm.go

export interface VmGet {
  VmID: string;
}
export interface VmDelete {
  VmID: string;
}
export interface VmUpdate {
  VmID: string;
}
export interface GpuAttach {
  VmID: string;
  GpuID: string;
}
export interface GpuDetach {
  VmID: string;
}
export interface GpuGet {
  GpuID: string;
}
export interface VmCommand {
  VmID: string;
}

//////////
// source: vm_snapshot.go

export interface VmSnapshotCreate {
  VmID: string;
}
export interface VmSnapshotList {
  VmID: string;
}
export interface VmSnapshotGet {
  VmID: string;
  SnapshotID: string;
}
export interface VmSnapshotDelete {
  VmID: string;
  SnapshotID: string;
}

//////////
// source: zone.go

export interface ZoneGet {
  Name: string;
  Type: string;
}
