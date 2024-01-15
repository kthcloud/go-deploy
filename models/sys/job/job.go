package job

import "time"

const (
	// TypeCreateVM is used when creating a VM.
	TypeCreateVM = "createVm"
	// TypeDeleteVM is used when deleting a VM.
	TypeDeleteVM = "deleteVm"
	// TypeUpdateVM is used when updating a VM.
	TypeUpdateVM = "updateVm"
	// TypeUpdateVmOwner is used when updating a VM's owner.
	TypeUpdateVmOwner = "updateVmOwner"
	// TypeAttachGPU is used when attaching a GPU to a VM.
	TypeAttachGPU = "attachGpu"
	// TypeDetachGPU is used when detaching a GPU from a VM.
	TypeDetachGPU = "detachGpu"
	// TypeRepairVM is used when repairing a VM.
	TypeRepairVM = "repairVm"
	// TypeCreateSystemSnapshot is used when creating a snapshot requested by the system.
	// This is separate from TypeCreateUserSnapshot because the system can create any number of snapshots.
	TypeCreateSystemSnapshot = "createSystemSnapshot"
	// TypeCreateUserSnapshot is used when creating a snapshot requested by a user.
	// This is separate from TypeCreateSystemSnapshot because the system can create any number of snapshots.
	TypeCreateUserSnapshot = "createUserSnapshot"
	// TypeDeleteSnapshot is used when deleting a snapshot.
	TypeDeleteSnapshot = "deleteSnapshot"

	// TypeCreateDeployment is used when creating a deployment.
	TypeCreateDeployment = "createDeployment"
	// TypeDeleteDeployment is used when deleting a deployment.
	TypeDeleteDeployment = "deleteDeployment"
	// TypeUpdateDeployment is used when updating a deployment.
	TypeUpdateDeployment = "updateDeployment"
	// TypeUpdateDeploymentOwner is used when updating a deployment's owner.
	TypeUpdateDeploymentOwner = "updateDeploymentOwner"
	// TypeBuildDeployments is used when building deployments.
	// This contains multiple deployments if they use the same image.
	TypeBuildDeployments = "buildDeployments"
	// TypeRepairDeployment is used when repairing a deployment.
	TypeRepairDeployment = "repairDeployment"

	// TypeCreateSM is used when creating a storage manager.
	TypeCreateSM = "createSM"
	// TypeDeleteSM is used when deleting a storage manager.
	TypeDeleteSM = "deleteSM"
	// TypeRepairSM is used when repairing a storage manager.
	TypeRepairSM = "repairSM"
)

const (
	// StatusPending is used when a job is pending and waiting to be run.
	StatusPending = "pending"
	// StatusRunning is used when a job is running.
	StatusRunning = "running"
	// StatusCompleted is used when a job is completed.
	StatusCompleted = "completed"
	// StatusFailed is used when a job has failed.
	StatusFailed = "failed"
	// StatusTerminated is used when a job has been terminated.
	StatusTerminated = "terminated"

	// StatusFinished
	// Deprecated: use StatusCompleted instead.
	StatusFinished = "finished"
)

type Job struct {
	ID     string                 `bson:"id" json:"id"`
	UserID string                 `bson:"userId" json:"userId"`
	Type   string                 `bson:"type" json:"type"`
	Args   map[string]interface{} `bson:"args" json:"args"`

	CreatedAt  time.Time `bson:"createdAt" json:"createdAt"`
	LastRunAt  time.Time `bson:"lastRunAt" json:"lastRunAt"`
	FinishedAt time.Time `bson:"finishedAt" json:"finishedAt"`
	RunAfter   time.Time `bson:"runAfter" json:"runAfter"`

	Attempts int `bson:"attempts" json:"attempts"`

	Status    string   `bson:"status" json:"status"`
	ErrorLogs []string `bson:"errorLogs" json:"errorLogs"`
}

type UpdateParams struct {
	Status *string `bson:"status" json:"status"`
}
