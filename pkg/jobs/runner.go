package jobs

import (
	"errors"
	"fmt"
	jobModel "go-deploy/models/sys/job"
	"go-deploy/utils"
	"strings"
)

type Runner struct {
	Job *jobModel.Job
}

func NewRunner(job *jobModel.Job) *Runner {
	return &Runner{Job: job}
}

func (runner *Runner) Run() {

	if jobDef := GetJobDef(runner.Job); jobDef != nil {
		if jobDef.TerminateFunc != nil {
			shouldTerminate, err := jobDef.TerminateFunc(runner.Job)
			if err != nil {
				err = jobModel.New().MarkTerminated(runner.Job.ID, err.Error())
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("error marking job as terminated. details: %w", err))
					return
				}
				return
			}

			if shouldTerminate {
				err = jobModel.New().MarkTerminated(runner.Job.ID, "gracefully terminated by system")
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("error marking job as terminated. details: %w", err))
					return
				}
				return
			}
		}

		go wrapper(jobDef.JobFunc, runner.Job)
	} else {
		utils.PrettyPrintError(fmt.Errorf("unknown job type: %s", runner.Job.Type))
	}
}

func wrapper(fn func(job *jobModel.Job) error, job *jobModel.Job) {
	err := fn(job)

	if err != nil {
		if strings.HasPrefix(err.Error(), "failed") {
			err = errors.Unwrap(err)
			utils.PrettyPrintError(fmt.Errorf("failed job (%s). details: %w", job.Type, err))

			err = jobModel.New().MarkFailed(job.ID, err.Error())
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("error marking job as failed. details: %w", err))
				return
			}
		} else if strings.HasPrefix(err.Error(), "terminated") {
			err = errors.Unwrap(err)
			utils.PrettyPrintError(fmt.Errorf("terminated job (%s). details: %w", job.Type, err))

			err = jobModel.New().MarkTerminated(job.ID, err.Error())
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("error marking job as terminated. details: %w", err))
				return
			}
		} else {
			utils.PrettyPrintError(fmt.Errorf("error executing job (%s). details: %w", job.Type, err))

			err = jobModel.New().MarkFailed(job.ID, err.Error())
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("error marking job as failed. details: %w", err))
				return
			}
		}
	} else {
		err = jobModel.New().MarkCompleted(job.ID)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("error marking job as completed. details: %w", err))
			return
		}
	}
}
