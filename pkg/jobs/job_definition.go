package jobs

import (
	"github.com/kthcloud/go-deploy/models/model"
	"github.com/kthcloud/go-deploy/models/version"
	"github.com/kthcloud/go-deploy/pkg/jobs/utils"
	v2 "github.com/kthcloud/go-deploy/pkg/jobs/v2"
)

// JobDefinition is a definition of a job.
// It contains the job itself and the functions that are executed when the job is created, updated, deleted, etc.
type JobDefinition struct {
	Job           *model.Job
	JobFunc       func(*model.Job) error
	EntryFunc     func(*model.Job) error
	ExitFunc      func(*model.Job) error
	TerminateFunc func(*model.Job) (bool, error)
}

type JobDefinitions map[string]JobDefinition

// GetJobDef returns the job definition for the given job.
func GetJobDef(job *model.Job) *JobDefinition {
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

	v2Defs := map[string]JobDefinition{
		// Deployment
		model.JobCreateDeployment: {
			JobFunc:       v2.CreateDeployment,
			TerminateFunc: coreJobDeployment.Build(),
			EntryFunc:     utils.DAddActivity(model.ActivityBeingCreated),
			ExitFunc:      utils.DRemActivity(model.ActivityBeingCreated),
		},
		model.JobDeleteDeployment: {
			JobFunc:   v2.DeleteDeployment,
			EntryFunc: utils.DAddActivity(model.ActivityBeingDeleted),
		},
		model.JobUpdateDeployment: {
			JobFunc:       v2.UpdateDeployment,
			TerminateFunc: leafJobDeployment.Build(),
			EntryFunc:     utils.DAddActivity(model.ActivityUpdating),
			ExitFunc:      utils.DRemActivity(model.ActivityUpdating),
		},
		model.JobUpdateDeploymentOwner: {
			JobFunc:       v2.UpdateDeploymentOwner,
			TerminateFunc: coreJobDeployment.Build(),
			EntryFunc:     utils.DAddActivity(model.ActivityUpdating),
			ExitFunc:      utils.DRemActivity(model.ActivityUpdating),
		},
		model.JobRepairDeployment: {
			JobFunc:       v2.RepairDeployment,
			TerminateFunc: leafJobDeployment.Build(),
			EntryFunc:     utils.DAddActivity(model.ActivityRepairing),
			ExitFunc:      utils.DRemActivity(model.ActivityRepairing),
		},

		// SM
		model.JobCreateSM: {
			JobFunc: v2.CreateSM,
		},
		model.JobDeleteSM: {
			JobFunc: v2.DeleteSM,
		},
		model.JobRepairSM: {
			JobFunc: v2.RepairSM,
		},

		// VM
		model.JobCreateVM: {
			JobFunc:       v2.CreateVM,
			TerminateFunc: coreJobVM.Build(),
			EntryFunc:     utils.VmAddActivity(model.ActivityBeingCreated),
			ExitFunc:      utils.VmRemActivity(model.ActivityBeingCreated),
		},
		model.JobDeleteVM: {
			JobFunc:   v2.DeleteVM,
			EntryFunc: utils.VmAddActivity(model.ActivityBeingDeleted),
		},
		model.JobUpdateVM: {
			JobFunc:       v2.UpdateVM,
			TerminateFunc: leafJobVM.Build(),
			EntryFunc:     utils.VmAddActivity(model.ActivityUpdating),
			ExitFunc:      utils.VmRemActivity(model.ActivityUpdating),
		},
		model.JobCreateGpuLease: {
			JobFunc:       v2.CreateGpuLease,
			TerminateFunc: leafJobVM.Build(),
		},
		model.JobUpdateGpuLease: {
			JobFunc:       v2.UpdateGpuLease,
			TerminateFunc: leafJobVM.Build(),
		},
		model.JobDeleteGpuLease: {
			JobFunc:       v2.DeleteGpuLease,
			TerminateFunc: leafJobVM.Build(),
		},
		model.JobCreateSystemVmSnapshot: {
			JobFunc:       v2.CreateSystemVmSnapshot,
			TerminateFunc: leafJobVM.Build(),
		},
		model.JobCreateVmUserSnapshot: {
			JobFunc:       v2.CreateUserVmSnapshot,
			TerminateFunc: oneCreateSnapshotPerUser.Build(),
		},
		model.JobDeleteVmSnapshot: {
			JobFunc:       v2.DeleteVmSnapshot,
			TerminateFunc: leafJobVM.Build(),
		},
		model.JobDoVmAction: {
			JobFunc:       v2.DoVmAction,
			TerminateFunc: leafJobVM.Build(),
		},
		model.JobUpdateVmOwner: {
			JobFunc:       v2.UpdateVmOwner,
			TerminateFunc: coreJobVM.Build(),
		},
		model.JobRepairVM: {
			JobFunc:       v2.RepairVM,
			TerminateFunc: leafJobVM.Build(),
		},

		// GpuClaim
		model.JobCreateGpuClaim: {
			JobFunc: v2.CreateGpuClaim,
		},
		model.JobDeleteGpuClaim: {
			JobFunc: v2.DeleteGpuClaim,
		},
		model.JobUpdateGpuClaim: {
			JobFunc: v2.UpdateGpuClaim,
		},
	}

	return map[string]map[string]JobDefinition{
		version.V2: v2Defs,
	}
}

// TerminateFuncBuilder is a builder for terminate functions.
// It uses the builder pattern to add multiple terminate functions to one terminate function.
type TerminateFuncBuilder struct {
	terminateFuncs []func(*model.Job) (bool, error)
}

// Builder returns a new TerminateFuncBuilder.
func Builder() *TerminateFuncBuilder {
	return &TerminateFuncBuilder{}
}

// Add adds a new terminate function to the builder.
func (builder *TerminateFuncBuilder) Add(terminateFunc func(*model.Job) (bool, error)) *TerminateFuncBuilder {
	builder.terminateFuncs = append(builder.terminateFuncs, terminateFunc)

	return builder
}

// Build builds the terminate function.
func (builder *TerminateFuncBuilder) Build() func(*model.Job) (bool, error) {
	return func(job *model.Job) (bool, error) {
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
