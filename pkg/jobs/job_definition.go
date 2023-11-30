package jobs

import (
	da "go-deploy/models/sys/deployment"
	jobModel "go-deploy/models/sys/job"
	va "go-deploy/models/sys/vm"
)

type JobDefinition struct {
	Job           *jobModel.Job
	JobFunc       func(*jobModel.Job) error
	EntryFunc     func(*jobModel.Job) error
	ExitFunc      func(*jobModel.Job) error
	TerminateFunc func(*jobModel.Job) (bool, error)
}

type JobDefinitions map[string]JobDefinition

func GetJobDef(job *jobModel.Job) *JobDefinition {
	jobDef, ok := jobMapper()[job.Type]
	if !ok {
		return nil
	}

	jobDef.Job = job

	return &jobDef
}

func jobMapper() map[string]JobDefinition {
	coreJobVM := Builder().Add(vmDeleted)
	leafJobVM := Builder().Add(vmDeleted).Add(updatingOwner)
	oneCreateSnapshotPerUser := Builder().Add(vmDeleted).Add(updatingOwner).Add(onlyCreateSnapshotPerUser)

	coreJobDeployment := Builder().Add(deploymentDeleted)
	leafJobDeployment := Builder().Add(deploymentDeleted).Add(updatingOwner)

	return map[string]JobDefinition{
		// vm
		jobModel.TypeCreateVM: {
			JobFunc:       CreateVM,
			TerminateFunc: coreJobVM.Build(),
		},
		jobModel.TypeDeleteVM: {
			JobFunc:   DeleteVM,
			EntryFunc: vAddActivity(va.ActivityBeingDeleted),
		},
		jobModel.TypeUpdateVM: {
			JobFunc:       UpdateVM,
			TerminateFunc: leafJobVM.Build(),
			EntryFunc:     vAddActivity(va.ActivityUpdating),
			ExitFunc:      vRemActivity(va.ActivityUpdating),
		},
		jobModel.TypeUpdateVmOwner: {
			JobFunc:       UpdateVmOwner,
			TerminateFunc: coreJobVM.Build(),
			EntryFunc:     vAddActivity(va.ActivityUpdating),
			ExitFunc:      vRemActivity(va.ActivityUpdating),
		},
		jobModel.TypeAttachGPU: {
			JobFunc:       AttachGpuToVM,
			TerminateFunc: leafJobVM.Build(),
			EntryFunc:     vAddActivity(va.ActivityAttachingGPU, va.ActivityUpdating),
			ExitFunc:      vRemActivity(va.ActivityAttachingGPU, va.ActivityUpdating),
		},
		jobModel.TypeDetachGPU: {
			JobFunc:       DetachGpuFromVM,
			TerminateFunc: leafJobVM.Build(),
			EntryFunc:     vAddActivity(va.ActivityDetachingGPU, va.ActivityUpdating),
			ExitFunc:      vRemActivity(va.ActivityDetachingGPU, va.ActivityUpdating),
		},
		jobModel.TypeRepairVM: {
			JobFunc:       RepairVM,
			TerminateFunc: leafJobVM.Build(),
			EntryFunc:     vAddActivity(va.ActivityRepairing),
			ExitFunc:      vRemActivity(va.ActivityRepairing),
		},
		jobModel.TypeCreateSystemSnapshot: {
			JobFunc:       CreateSystemSnapshot,
			TerminateFunc: leafJobVM.Build(),
			EntryFunc:     vAddActivity(va.ActivityCreatingSnapshot),
			ExitFunc:      vRemActivity(va.ActivityCreatingSnapshot),
		},
		jobModel.TypeCreateUserSnapshot: {
			JobFunc:       CreateUserSnapshot,
			TerminateFunc: oneCreateSnapshotPerUser.Build(),
			EntryFunc:     vAddActivity(va.ActivityCreatingSnapshot),
			ExitFunc:      vRemActivity(va.ActivityCreatingSnapshot),
		},
		jobModel.TypeDeleteSnapshot: {
			JobFunc:       DeleteSnapshot,
			TerminateFunc: leafJobVM.Build(),
		},
		jobModel.TypeApplySnapshot: {
			JobFunc:       ApplySnapshot,
			TerminateFunc: leafJobVM.Build(),
			EntryFunc:     vAddActivity(va.ActivityApplyingSnapshot),
			ExitFunc:      vRemActivity(va.ActivityApplyingSnapshot),
		},

		// deployment
		jobModel.TypeCreateDeployment: {
			JobFunc:       CreateDeployment,
			TerminateFunc: coreJobDeployment.Build(),
		},
		jobModel.TypeDeleteDeployment: {
			JobFunc:   DeleteDeployment,
			EntryFunc: dAddActivity(da.ActivityBeingDeleted),
		},
		jobModel.TypeUpdateDeployment: {
			JobFunc:       UpdateDeployment,
			TerminateFunc: leafJobDeployment.Build(),
			EntryFunc:     dAddActivity(da.ActivityUpdating),
			ExitFunc:      dRemActivity(da.ActivityUpdating),
		},
		jobModel.TypeUpdateDeploymentOwner: {
			JobFunc:       UpdateDeploymentOwner,
			TerminateFunc: coreJobDeployment.Build(),
			EntryFunc:     dAddActivity(da.ActivityUpdating),
			ExitFunc:      dRemActivity(da.ActivityUpdating),
		},
		jobModel.TypeBuildDeployments: {
			// this is a special case where multiple deployments are built in one job, so we don't want to terminate it
			JobFunc: BuildDeployments,
		},
		jobModel.TypeRepairDeployment: {
			JobFunc:       RepairDeployment,
			TerminateFunc: leafJobDeployment.Build(),
			EntryFunc:     dAddActivity(da.ActivityRepairing),
			ExitFunc:      dRemActivity(da.ActivityRepairing),
		},

		// storage manager
		jobModel.TypeCreateStorageManager: {
			JobFunc: CreateStorageManager,
		},
		jobModel.TypeDeleteStorageManager: {
			JobFunc: DeleteStorageManager,
		},
		jobModel.TypeRepairStorageManager: {
			JobFunc: RepairStorageManager,
		},
	}
}

type TerminateFuncBuilder struct {
	terminateFuncs []func(*jobModel.Job) (bool, error)
}

func Builder() *TerminateFuncBuilder {
	return &TerminateFuncBuilder{}
}

func (builder *TerminateFuncBuilder) Add(terminateFunc func(*jobModel.Job) (bool, error)) *TerminateFuncBuilder {
	builder.terminateFuncs = append(builder.terminateFuncs, terminateFunc)

	return builder
}

func (builder *TerminateFuncBuilder) Build() func(*jobModel.Job) (bool, error) {
	return func(job *jobModel.Job) (bool, error) {
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
