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
)

const (
	StatusPending  = "pending"
	StatusRunning  = "running"
	StatusFinished = "finished"
	StatusFailed   = "failed"
)

type Job struct {
	ID        string                 `bson:"id" json:"id"`
	UserID    string                 `bson:"userId" json:"userId"`
	Type      string                 `bson:"type" json:"type"`
	Args      map[string]interface{} `bson:"args" json:"args"`
	CreatedAt time.Time              `bson:"createdAt" json:"createdAt"`
	Status    string                 `bson:"status" json:"status"`
	ErrorLogs []string               `bson:"errorLogs" json:"errorLogs"`
}
