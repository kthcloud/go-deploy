package model

import "time"

const (
	// JobCreateVM is used when creating a VM.
	JobCreateVM = "createVm"
	// JobDeleteVM is used when deleting a VM.
	JobDeleteVM = "deleteVm"
	// JobUpdateVM is used when updating a VM.
	JobUpdateVM = "updateVm"
	// JobUpdateVmOwner is used when updating a VM's owner.
	JobUpdateVmOwner = "updateVmOwner"
	// JobAttachGPU is used when attaching a GPU to a VM.
	JobAttachGPU = "attachGpu"
	// JobDetachGPU is used when detaching a GPU from a VM.
	JobDetachGPU = "detachGpu"
	// JobRepairVM is used when repairing a VM.
	JobRepairVM = "repairVm"
	// JobCreateSystemVmSnapshot is used when creating a snapshot requested by the system.
	// This is separate from JobCreateVmUserSnapshot because the system can create any number of snapshots.
	JobCreateSystemVmSnapshot = "createSystemSnapshot"
	// JobCreateVmUserSnapshot is used when creating a snapshot requested by a user.
	// This is separate from JobCreateSystemVmSnapshot because the system can create any number of snapshots.
	JobCreateVmUserSnapshot = "createUserSnapshot"
	// JobDeleteVmSnapshot is used when deleting a snapshot.
	JobDeleteVmSnapshot = "deleteSnapshot"
	// JobDoVmAction is used when doing an action on a VM.
	JobDoVmAction = "doVmAction"

	// JobCreateDeployment is used when creating a deployment.
	JobCreateDeployment = "createDeployment"
	// JobDeleteDeployment is used when deleting a deployment.
	JobDeleteDeployment = "deleteDeployment"
	// JobUpdateDeployment is used when updating a deployment.
	JobUpdateDeployment = "updateDeployment"
	// JobUpdateDeploymentOwner is used when updating a deployment's owner.
	JobUpdateDeploymentOwner = "updateDeploymentOwner"
	// JobRepairDeployment is used when repairing a deployment.
	JobRepairDeployment = "repairDeployment"

	// JobCreateSM is used when creating a storage manager.
	JobCreateSM = "createSm"
	// JobDeleteSM is used when deleting a storage manager.
	JobDeleteSM = "deleteSm"
	// JobRepairSM is used when repairing a storage manager.
	JobRepairSM = "repairSm"
)

const (
	// JobStatusPending is used when a job is pending and waiting to be run.
	JobStatusPending = "pending"
	// JobStatusRunning is used when a job is running.
	JobStatusRunning = "running"
	// JobStatusCompleted is used when a job is completed.
	JobStatusCompleted = "completed"
	// JobStatusFailed is used when a job has failed.
	JobStatusFailed = "failed"
	// JobStatusTerminated is used when a job has been terminated.
	JobStatusTerminated = "terminated"
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

type JobUpdateParams struct {
	Status *string `bson:"status" json:"status"`
}
