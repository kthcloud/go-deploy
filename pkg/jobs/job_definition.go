package jobs

import (
	da "go-deploy/models/sys/deployment"
	jobModels "go-deploy/models/sys/job"
	sa "go-deploy/models/sys/sm"
	va "go-deploy/models/sys/vm"
)

// JobDefinition is a definition of a job.
// It contains the job itself and the functions that are executed when the job is created, updated, deleted, etc.
type JobDefinition struct {
	Job           *jobModels.Job
	JobFunc       func(*jobModels.Job) error
	EntryFunc     func(*jobModels.Job) error
	ExitFunc      func(*jobModels.Job) error
	TerminateFunc func(*jobModels.Job) (bool, error)
}

type JobDefinitions map[string]JobDefinition

// GetJobDef returns the job definition for the given job.
func GetJobDef(job *jobModels.Job) *JobDefinition {
	jobDef, ok := jobMapper()[job.Type]
	if !ok {
		return nil
	}

	jobDef.Job = job

	return &jobDef
}

// jobMapper maps job types to job definitions.
func jobMapper() map[string]JobDefinition {
	coreJobVM := Builder().Add(vmDeleted)
	leafJobVM := Builder().Add(vmDeleted).Add(updatingOwner)
	oneCreateSnapshotPerUser := Builder().Add(vmDeleted).Add(updatingOwner).Add(onlyCreateSnapshotPerUser)

	coreJobDeployment := Builder().Add(deploymentDeleted)
	leafJobDeployment := Builder().Add(deploymentDeleted).Add(updatingOwner)

	return map[string]JobDefinition{
		// vm
		jobModels.TypeCreateVM: {
			JobFunc:       CreateVM,
			TerminateFunc: coreJobVM.Build(),
			EntryFunc:     vAddActivity(va.ActivityBeingCreated),
			ExitFunc:      vRemActivity(va.ActivityBeingCreated),
		},
		jobModels.TypeDeleteVM: {
			JobFunc:   DeleteVM,
			EntryFunc: vAddActivity(va.ActivityBeingDeleted),
		},
		jobModels.TypeUpdateVM: {
			JobFunc:       UpdateVM,
			TerminateFunc: leafJobVM.Build(),
			EntryFunc:     vAddActivity(va.ActivityUpdating),
			ExitFunc:      vRemActivity(va.ActivityUpdating),
		},
		jobModels.TypeUpdateVmOwner: {
			JobFunc:       UpdateVmOwner,
			TerminateFunc: coreJobVM.Build(),
			EntryFunc:     vAddActivity(va.ActivityUpdating),
			ExitFunc:      vRemActivity(va.ActivityUpdating),
		},
		jobModels.TypeAttachGPU: {
			JobFunc:       AttachGpuToVM,
			TerminateFunc: leafJobVM.Build(),
			EntryFunc:     vAddActivity(va.ActivityAttachingGPU, va.ActivityUpdating),
			ExitFunc:      vRemActivity(va.ActivityAttachingGPU, va.ActivityUpdating),
		},
		jobModels.TypeDetachGPU: {
			JobFunc:       DetachGpuFromVM,
			TerminateFunc: leafJobVM.Build(),
			EntryFunc:     vAddActivity(va.ActivityDetachingGPU, va.ActivityUpdating),
			ExitFunc:      vRemActivity(va.ActivityDetachingGPU, va.ActivityUpdating),
		},
		jobModels.TypeRepairVM: {
			JobFunc:       RepairVM,
			TerminateFunc: leafJobVM.Build(),
			EntryFunc:     vAddActivity(va.ActivityRepairing),
			ExitFunc:      vRemActivity(va.ActivityRepairing),
		},
		jobModels.TypeCreateSystemSnapshot: {
			JobFunc:       CreateSystemSnapshot,
			TerminateFunc: leafJobVM.Build(),
		},
		jobModels.TypeCreateUserSnapshot: {
			JobFunc:       CreateUserSnapshot,
			TerminateFunc: oneCreateSnapshotPerUser.Build(),
		},
		jobModels.TypeDeleteSnapshot: {
			JobFunc:       DeleteSnapshot,
			TerminateFunc: leafJobVM.Build(),
		},

		// deployment
		jobModels.TypeCreateDeployment: {
			JobFunc:       CreateDeployment,
			TerminateFunc: coreJobDeployment.Build(),
			EntryFunc:     dAddActivity(da.ActivityBeingCreated),
			ExitFunc:      dRemActivity(da.ActivityBeingCreated),
		},
		jobModels.TypeDeleteDeployment: {
			JobFunc:   DeleteDeployment,
			EntryFunc: dAddActivity(da.ActivityBeingDeleted),
		},
		jobModels.TypeUpdateDeployment: {
			JobFunc:       UpdateDeployment,
			TerminateFunc: leafJobDeployment.Build(),
			EntryFunc:     dAddActivity(da.ActivityUpdating),
			ExitFunc:      dRemActivity(da.ActivityUpdating),
		},
		jobModels.TypeUpdateDeploymentOwner: {
			JobFunc:       UpdateDeploymentOwner,
			TerminateFunc: coreJobDeployment.Build(),
			EntryFunc:     dAddActivity(da.ActivityUpdating),
			ExitFunc:      dRemActivity(da.ActivityUpdating),
		},
		jobModels.TypeBuildDeployments: {
			// this is a special case where multiple deployments are built in one job, so we don't want to terminate it
			JobFunc: BuildDeployments,
		},
		jobModels.TypeRepairDeployment: {
			JobFunc:       RepairDeployment,
			TerminateFunc: leafJobDeployment.Build(),
			EntryFunc:     dAddActivity(da.ActivityRepairing),
			ExitFunc:      dRemActivity(da.ActivityRepairing),
		},

		// storage manager
		jobModels.TypeCreateSM: {
			JobFunc:   CreateSM,
			EntryFunc: sAddActivity(sa.ActivityBeingCreated),
			ExitFunc:  sRemActivity(sa.ActivityBeingCreated),
		},
		jobModels.TypeDeleteSM: {
			JobFunc:   DeleteSM,
			EntryFunc: sAddActivity(sa.ActivityBeingDeleted),
		},
		jobModels.TypeRepairSM: {
			JobFunc:   RepairSM,
			EntryFunc: sAddActivity(sa.ActivityBeingCreated),
			ExitFunc:  sRemActivity(sa.ActivityRepairing),
		},
	}
}

// TerminateFuncBuilder is a builder for terminate functions.
// It uses the builder pattern to add multiple terminate functions to one terminate function.
type TerminateFuncBuilder struct {
	terminateFuncs []func(*jobModels.Job) (bool, error)
}

// Builder returns a new TerminateFuncBuilder.
func Builder() *TerminateFuncBuilder {
	return &TerminateFuncBuilder{}
}

// Add adds a new terminate function to the builder.
func (builder *TerminateFuncBuilder) Add(terminateFunc func(*jobModels.Job) (bool, error)) *TerminateFuncBuilder {
	builder.terminateFuncs = append(builder.terminateFuncs, terminateFunc)

	return builder
}

// Build builds the terminate function.
func (builder *TerminateFuncBuilder) Build() func(*jobModels.Job) (bool, error) {
	return func(job *jobModels.Job) (bool, error) {
		for _, terminateFunc := range builder.terminateFuncs {
			res, err := terminateFunc(job)
			if err != nil {
				return false, err
			}

			if res {
				return true, nil
			}
		}

		return false, nil
	}
}
