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
	// TypeCreateSystemVmSnapshot is used when creating a snapshot requested by the system.
	// This is separate from TypeCreateVmUserSnapshot because the system can create any number of snapshots.
	TypeCreateSystemVmSnapshot = "createSystemSnapshot"
	// TypeCreateVmUserSnapshot is used when creating a snapshot requested by a user.
	// This is separate from TypeCreateSystemVmSnapshot because the system can create any number of snapshots.
	TypeCreateVmUserSnapshot = "createUserSnapshot"
	// TypeDeleteVmSnapshot is used when deleting a snapshot.
	TypeDeleteVmSnapshot = "deleteSnapshot"
	// TypeDoVmAction is used when doing an action on a VM.
	TypeDoVmAction = "doVmAction"

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
	TypeCreateSM = "createSm"
	// TypeDeleteSM is used when deleting a storage manager.
	TypeDeleteSM = "deleteSm"
	// TypeRepairSM is used when repairing a storage manager.
	TypeRepairSM = "repairSm"
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
	ID      string                 `bson:"id"`
	UserID  string                 `bson:"userId"`
	Type    string                 `bson:"type"`
	Args    map[string]interface{} `bson:"args"`
	Version string                 `bson:"version"`

	CreatedAt  time.Time `bson:"createdAt"`
	LastRunAt  time.Time `bson:"lastRunAt,omitempty"`
	FinishedAt time.Time `bson:"finishedAt,omitempty"`
	RunAfter   time.Time `bson:"runAfter,omitempty"`

	Attempts int `bson:"attempts"`

	Status    string   `bson:"status" `
	ErrorLogs []string `bson:"errorLogs" `
}

type UpdateParams struct {
	Status *string `bson:"status" json:"status"`
}
