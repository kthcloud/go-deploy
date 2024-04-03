package jobs

import (
	"errors"
	"fmt"
	"go-deploy/models/model"
	"go-deploy/pkg/db/resources/job_repo"
	"go-deploy/pkg/log"
	"go-deploy/utils"
	"math"
	"strings"
	"time"
)

type Runner struct {
	Job *model.Job
}

const jobAttemptsLimit = 5

// NewRunner creates a new job runner for the given job.
func NewRunner(job *model.Job) *Runner {
	return &Runner{Job: job}
}

// Run runs the job, and marks it as completed, failed or terminated according to the job's result.
func (runner *Runner) Run() {
	if jobDef := GetJobDef(runner.Job); jobDef != nil {
		if jobDef.TerminateFunc != nil {
			shouldTerminate, err := jobDef.TerminateFunc(runner.Job)
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("error executing job %s (%s) terminate function, terminating the job instead. details: %w", runner.Job.ID, runner.Job.Type, err))

				err = job_repo.New().MarkTerminated(runner.Job.ID, err.Error())
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("error marking job %s (%s) as terminated. details: %w", runner.Job.ID, runner.Job.Type, err))
					return
				}
				return
			}

			if shouldTerminate {
				log.Println("Job %s (%s) gracefully terminated by system", runner.Job.ID, runner.Job.Type)
				err = job_repo.New().MarkTerminated(runner.Job.ID, "gracefully terminated by system")
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("error marking job %s (%s) as terminated. details: %w", runner.Job.ID, runner.Job.Type, err))
					return
				}
				return
			}
		}

		go wrapper(jobDef)
	} else {
		utils.PrettyPrintError(fmt.Errorf("job %s has unknown type %s", runner.Job.ID, runner.Job.Type))

		err := job_repo.New().MarkTerminated(runner.Job.ID, fmt.Sprintf("job %s has unknown type %s", runner.Job.ID, runner.Job.Type))
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("error marking unknown job %s (%s) as terminated. details: %w", runner.Job.ID, runner.Job.Type, err))
			return
		}
	}
}

// wrapper is a helper function that runs the EntryFunc, JobFunc and ExitFunc of the given job definition,
// and updates the job's status according to the result of the JobFunc.
func wrapper(def *JobDefinition) {
	if def.EntryFunc != nil {
		err := def.EntryFunc(def.Job)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("error executing job %s (%s) entry function, terminating. details: %w", def.Job.ID, def.Job.Type, err))

			err = job_repo.New().MarkTerminated(def.Job.ID, err.Error())
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("error marking job as terminated. details: %w", err))
				return
			}
			return
		}
	}

	defer func() {
		if def.ExitFunc != nil {
			err := def.ExitFunc(def.Job)
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("error executing job %s (%s) exit function, terminating. details: %w", def.Job.ID, def.Job.Type, err))

				err = job_repo.New().MarkTerminated(def.Job.ID, err.Error())
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("error marking job as failed. details: %w", err))
					return
				}
				return
			}
		}
	}()

	err := def.JobFunc(def.Job)

	if err != nil {
		if strings.HasPrefix(err.Error(), "failed") {
			err = errors.Unwrap(err)

			attempts := def.Job.Attempts + 1
			if attempts >= jobAttemptsLimit {
				utils.PrettyPrintError(fmt.Errorf("terminated job %s (%s) after %d failed attempts. details: %w", def.Job.ID, def.Job.Type, attempts, err))
				err = job_repo.New().MarkTerminated(def.Job.ID, err.Error())
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("error marking job %s (%s) as terminated. details: %w", def.Job.ID, def.Job.Type, err))
					return
				}
				return
			} else {
				delay := int(math.Pow(2, float64(attempts-1)) * 30)
				runAfter := def.Job.RunAfter.Add(time.Duration(delay) * time.Second)
				utils.PrettyPrintError(fmt.Errorf("failed job %s (%s), attempt: %d/%d delay: %ds details: %w", def.Job.ID, def.Job.Type, attempts, jobAttemptsLimit, delay, err))

				err = job_repo.New().MarkFailed(def.Job.ID, runAfter, attempts, err.Error())
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("error marking job %s (%s) as failed. details: %w", def.Job.ID, def.Job.Type, err))
					return
				}
			}
		} else if strings.HasPrefix(err.Error(), "terminated") {
			err = errors.Unwrap(err)
			utils.PrettyPrintError(fmt.Errorf("terminated job %s (%s). details: %w", def.Job.ID, def.Job.Type, err))

			err = job_repo.New().MarkTerminated(def.Job.ID, err.Error())
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("error marking job %s (%s) as terminated. details: %w", def.Job.ID, def.Job.Type, err))
				return
			}
		} else {
			utils.PrettyPrintError(fmt.Errorf("error executing job %s (%s), terminating. details: %w", def.Job.ID, def.Job.Type, err))

			err = job_repo.New().MarkTerminated(def.Job.ID, err.Error())
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("error marking job %s (%s) as terminated. details: %w", def.Job.ID, def.Job.Type, err))
				return
			}
		}
	} else {
		err = job_repo.New().MarkCompleted(def.Job.ID)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("error marking job %s (%s) as completed. details: %w", def.Job.ID, def.Job.Type, err))
			return
		}
	}
}
