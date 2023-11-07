package job

import "time"

const (
	TypeCreateVM              = "createVm"
	TypeDeleteVM              = "deleteVm"
	TypeUpdateVM              = "updateVm"
	TypeUpdateVmOwner         = "updateVmOwner"
	TypeAttachGpuToVM         = "attachGpuToVm"
	TypeDetachGpuFromVM       = "detachGpuFromVm"
	TypeCreateDeployment      = "createDeployment"
	TypeDeleteDeployment      = "deleteDeployment"
	TypeUpdateDeployment      = "updateDeployment"
	TypeUpdateDeploymentOwner = "updateDeploymentOwner"
	TypeBuildDeployments      = "buildDeployments"
	TypeCreateStorageManager  = "createStorageManager"
	TypeDeleteStorageManager  = "deleteStorageManager"
	TypeRepairStorageManager  = "repairStorageManager"
	TypeRepairVM              = "repairVm"
	TypeRepairDeployment      = "repairDeployment"
	TypeRepairGPUs            = "repairGpus"
	TypeCreateSnapshot        = "createSnapshot"
	TypeApplySnapshot         = "applySnapshot"
)

const (
	StatusPending    = "pending"
	StatusRunning    = "running"
	StatusCompleted  = "completed"
	StatusFailed     = "failed"
	StatusTerminated = "terminated"

	// StatusFinished deprecated
	StatusFinished = "finished"
)

type Job struct {
	ID          string                 `bson:"id" json:"id"`
	UserID      string                 `bson:"userId" json:"userId"`
	Type        string                 `bson:"type" json:"type"`
	Args        map[string]interface{} `bson:"args" json:"args"`
	ResourceIDs []string               `bson:"resourceIds" json:"resourceIds"`
	CreatedAt   time.Time              `bson:"createdAt" json:"createdAt"`
	LastRunAt   time.Time              `bson:"lastRunAt" json:"lastRunAt"`
	FinishedAt  time.Time              `bson:"finishedAt" json:"finishedAt"`
	RunAfter    time.Time              `bson:"runAfter" json:"runAfter"`
	Status      string                 `bson:"status" json:"status"`
	ErrorLogs   []string               `bson:"errorLogs" json:"errorLogs"`
}

type UpdateParams struct {
	Status *string `bson:"status" json:"status"`
}
