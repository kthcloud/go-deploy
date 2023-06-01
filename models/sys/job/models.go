package job

import "time"

const (
	TypeCreateVM         = "createVm"
	TypeDeleteVM         = "deleteVm"
	TypeUpdateVM         = "updateVm"
	TypeAttachGpuToVM    = "attachGpuToVm"
	TypeDetachGpuFromVM  = "detachGpuFromVm"
	TypeCreateDeployment = "createDeployment"
	TypeDeleteDeployment = "deleteDeployment"
	TypeUpdateDeployment = "updateDeployment"
	TypeBuildDeployment  = "buildDeployment"
	TypeRepairVM         = "repairVm"
	TypeRepairDeployment = "repairDeployment"
	TypeRepairGPUs       = "repairGpus"
)

const (
	StatusPending    = "pending"
	StatusRunning    = "running"
	StatusFinished   = "finished"
	StatusFailed     = "failed"
	StatusTerminated = "terminated"
)

type Job struct {
	ID        string                 `bson:"id" json:"id"`
	UserID    string                 `bson:"userId" json:"userId"`
	Type      string                 `bson:"type" json:"type"`
	Args      map[string]interface{} `bson:"args" json:"args"`
	CreatedAt time.Time              `bson:"createdAt" json:"createdAt"`
	LastRunAt time.Time              `bson:"lastRunAt" json:"lastRunAt"`
	Status    string                 `bson:"status" json:"status"`
	ErrorLogs []string               `bson:"errorLogs" json:"errorLogs"`
}
