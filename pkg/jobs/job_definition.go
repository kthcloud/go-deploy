package jobs

import (
	da "go-deploy/models/sys/deployment"
	jobModels "go-deploy/models/sys/job"
	sa "go-deploy/models/sys/sm"
	va "go-deploy/models/sys/vm"
	"go-deploy/models/versions"
	"go-deploy/pkg/jobs/utils"
	"go-deploy/pkg/jobs/v1"
	v2 "go-deploy/pkg/jobs/v2"
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
	jobDef, ok := jobMapper()[job.Version][job.Type]
	if !ok {
		return nil
	}

	jobDef.Job = job

	return &jobDef
}

// jobMapper maps job types to job definitions.
func jobMapper() map[string]map[string]JobDefinition {
	coreJobVM := Builder().Add(utils.VmDeleted)
	leafJobVM := Builder().Add(utils.VmDeleted).Add(utils.UpdatingOwner)
	oneCreateSnapshotPerUser := Builder().Add(utils.VmDeleted).Add(utils.UpdatingOwner).Add(utils.OnlyCreateSnapshotPerUser)

	coreJobDeployment := Builder().Add(utils.DeploymentDeleted)
	leafJobDeployment := Builder().Add(utils.DeploymentDeleted).Add(utils.UpdatingOwner)

	v1Defs := map[string]JobDefinition{
		// VM
		jobModels.TypeCreateVM: {
			JobFunc:       v1.CreateVM,
			TerminateFunc: coreJobVM.Build(),
			EntryFunc:     utils.VmAddActivity(va.ActivityBeingCreated),
			ExitFunc:      utils.VmRemActivity(va.ActivityBeingCreated),
		},
		jobModels.TypeDeleteVM: {
			JobFunc:   v1.DeleteVM,
			EntryFunc: utils.VmAddActivity(va.ActivityBeingDeleted),
		},
		jobModels.TypeUpdateVM: {
			JobFunc:       v1.UpdateVM,
			TerminateFunc: leafJobVM.Build(),
			EntryFunc:     utils.VmAddActivity(va.ActivityUpdating),
			ExitFunc:      utils.VmRemActivity(va.ActivityUpdating),
		},
		jobModels.TypeUpdateVmOwner: {
			JobFunc:       v1.UpdateVmOwner,
			TerminateFunc: coreJobVM.Build(),
			EntryFunc:     utils.VmAddActivity(va.ActivityUpdating),
			ExitFunc:      utils.VmRemActivity(va.ActivityUpdating),
		},
		jobModels.TypeAttachGPU: {
			JobFunc:       v1.AttachGpuToVM,
			TerminateFunc: leafJobVM.Build(),
			EntryFunc:     utils.VmAddActivity(va.ActivityAttachingGPU, va.ActivityUpdating),
			ExitFunc:      utils.VmRemActivity(va.ActivityAttachingGPU, va.ActivityUpdating),
		},
		jobModels.TypeDetachGPU: {
			JobFunc:       v1.DetachGpuFromVM,
			TerminateFunc: leafJobVM.Build(),
			EntryFunc:     utils.VmAddActivity(va.ActivityDetachingGPU, va.ActivityUpdating),
			ExitFunc:      utils.VmRemActivity(va.ActivityDetachingGPU, va.ActivityUpdating),
		},
		jobModels.TypeRepairVM: {
			JobFunc:       v1.RepairVM,
			TerminateFunc: leafJobVM.Build(),
			EntryFunc:     utils.VmAddActivity(va.ActivityRepairing),
			ExitFunc:      utils.VmRemActivity(va.ActivityRepairing),
		},
		jobModels.TypeCreateSystemVmSnapshot: {
			JobFunc:       v1.CreateSystemSnapshot,
			TerminateFunc: leafJobVM.Build(),
		},
		jobModels.TypeCreateVmUserSnapshot: {
			JobFunc:       v1.CreateUserSnapshot,
			TerminateFunc: oneCreateSnapshotPerUser.Build(),
		},
		jobModels.TypeDeleteVmSnapshot: {
			JobFunc:       v1.DeleteSnapshot,
			TerminateFunc: leafJobVM.Build(),
		},

		// Deployment
		jobModels.TypeCreateDeployment: {
			JobFunc:       v1.CreateDeployment,
			TerminateFunc: coreJobDeployment.Build(),
			EntryFunc:     utils.DAddActivity(da.ActivityBeingCreated),
			ExitFunc:      utils.DRemActivity(da.ActivityBeingCreated),
		},
		jobModels.TypeDeleteDeployment: {
			JobFunc:   v1.DeleteDeployment,
			EntryFunc: utils.DAddActivity(da.ActivityBeingDeleted),
		},
		jobModels.TypeUpdateDeployment: {
			JobFunc:       v1.UpdateDeployment,
			TerminateFunc: leafJobDeployment.Build(),
			EntryFunc:     utils.DAddActivity(da.ActivityUpdating),
			ExitFunc:      utils.DRemActivity(da.ActivityUpdating),
		},
		jobModels.TypeUpdateDeploymentOwner: {
			JobFunc:       v1.UpdateDeploymentOwner,
			TerminateFunc: coreJobDeployment.Build(),
			EntryFunc:     utils.DAddActivity(da.ActivityUpdating),
			ExitFunc:      utils.DRemActivity(da.ActivityUpdating),
		},
		jobModels.TypeBuildDeployments: {
			// This is a special case where multiple deployments are built in one job, so we don't want to terminate it
			JobFunc: v1.BuildDeployments,
		},
		jobModels.TypeRepairDeployment: {
			JobFunc:       v1.RepairDeployment,
			TerminateFunc: leafJobDeployment.Build(),
			EntryFunc:     utils.DAddActivity(da.ActivityRepairing),
			ExitFunc:      utils.DRemActivity(da.ActivityRepairing),
		},

		// storage manager
		jobModels.TypeCreateSM: {
			JobFunc:   v1.CreateSM,
			EntryFunc: utils.SmAddActivity(sa.ActivityBeingCreated),
			ExitFunc:  utils.SmRemActivity(sa.ActivityBeingCreated),
		},
		jobModels.TypeDeleteSM: {
			JobFunc:   v1.DeleteSM,
			EntryFunc: utils.SmAddActivity(sa.ActivityBeingDeleted),
		},
		jobModels.TypeRepairSM: {
			JobFunc:   v1.RepairSM,
			EntryFunc: utils.SmAddActivity(sa.ActivityBeingCreated),
			ExitFunc:  utils.SmRemActivity(sa.ActivityRepairing),
		},
	}

	v2Defs := map[string]JobDefinition{
		// VM
		jobModels.TypeCreateVM: {
			JobFunc:       v2.CreateVM,
			TerminateFunc: coreJobVM.Build(),
			EntryFunc:     utils.VmAddActivity(va.ActivityBeingCreated),
			ExitFunc:      utils.VmRemActivity(va.ActivityBeingCreated),
		},
		jobModels.TypeDeleteVM: {
			JobFunc:   v2.DeleteVM,
			EntryFunc: utils.VmAddActivity(va.ActivityBeingDeleted),
		},
		jobModels.TypeUpdateVM: {
			JobFunc:       v2.UpdateVM,
			TerminateFunc: leafJobVM.Build(),
			EntryFunc:     utils.VmAddActivity(va.ActivityUpdating),
			ExitFunc:      utils.VmRemActivity(va.ActivityUpdating),
		},
		//jobModels.TypeUpdateVmOwner: {
		//	JobFunc:       v2.UpdateVmOwner,
		//	TerminateFunc: coreJobVM.Build(),
		//	EntryFunc:     vAddActivity(va.ActivityUpdating),
		//	ExitFunc:      vRemActivity(va.ActivityUpdating),
		//},
		//jobModels.TypeAttachGPU: {
		//	JobFunc:       v2.AttachGpuToVM,
		//	TerminateFunc: leafJobVM.Build(),
		//	EntryFunc:     vAddActivity(va.ActivityAttachingGPU, va.ActivityUpdating),
		//	ExitFunc:      vRemActivity(va.ActivityAttachingGPU, va.ActivityUpdating),
		//},
		//jobModels.TypeDetachGPU: {
		//	JobFunc:       v2.DetachGpuFromVM,
		//	TerminateFunc: leafJobVM.Build(),
		//	EntryFunc:     vAddActivity(va.ActivityDetachingGPU, va.ActivityUpdating),
		//	ExitFunc:      vRemActivity(va.ActivityDetachingGPU, va.ActivityUpdating),
		//},
		//jobModels.TypeRepairVM: {
		//	JobFunc:       v2.RepairVM,
		//	TerminateFunc: leafJobVM.Build(),
		//	EntryFunc:     vAddActivity(va.ActivityRepairing),
		//	ExitFunc:      vRemActivity(va.ActivityRepairing),
		//},
		jobModels.TypeCreateSystemVmSnapshot: {
			JobFunc:       v2.CreateSystemVmSnapshot,
			TerminateFunc: leafJobVM.Build(),
		},
		jobModels.TypeCreateVmUserSnapshot: {
			JobFunc:       v2.CreateUserVmSnapshot,
			TerminateFunc: oneCreateSnapshotPerUser.Build(),
		},
		jobModels.TypeDeleteVmSnapshot: {
			JobFunc:       v2.DeleteVmSnapshot,
			TerminateFunc: leafJobVM.Build(),
		},
		jobModels.TypeDoVmAction: {
			JobFunc:       v2.DoVmAction,
			TerminateFunc: leafJobVM.Build(),
		},
	}

	return map[string]map[string]JobDefinition{
		versions.V1: v1Defs,
		versions.V2: v2Defs,
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
