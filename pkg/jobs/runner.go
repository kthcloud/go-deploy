package jobs

import (
	"errors"
	"fmt"
	jobModels "go-deploy/models/sys/job"
	"go-deploy/utils"
	"math"
	"strings"
	"time"
)

type Runner struct {
	Job *jobModels.Job
}

const jobAttemptsLimit = 5

func NewRunner(job *jobModels.Job) *Runner {
	return &Runner{Job: job}
}

func (runner *Runner) Run() {
	if jobDef := GetJobDef(runner.Job); jobDef != nil {
		if jobDef.TerminateFunc != nil {
			shouldTerminate, err := jobDef.TerminateFunc(runner.Job)
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("error executing job %s (%s) terminate function, terminating the job instead. details: %w", runner.Job.ID, runner.Job.Type, err))

				err = jobModels.New().MarkTerminated(runner.Job.ID, err.Error())
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("error marking job %s (%s) as terminated. details: %w", runner.Job.ID, runner.Job.Type, err))
					return
				}
				return
			}

			if shouldTerminate {
				err = jobModels.New().MarkTerminated(runner.Job.ID, "gracefully terminated by system")
				utils.PrettyPrintError(fmt.Errorf("job %s (%s) gracefully terminated by system", runner.Job.ID, runner.Job.Type))
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

		err := jobModels.New().MarkTerminated(runner.Job.ID, fmt.Sprintf("job %s has unknown type %s", runner.Job.ID, runner.Job.Type))
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("error marking unknown job %s (%s) as terminated. details: %w", runner.Job.ID, runner.Job.Type, err))
			return
		}
	}
}

func wrapper(def *JobDefinition) {
	if def.EntryFunc != nil {
		err := def.EntryFunc(def.Job)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("error executing job %s (%s) entry function, terminating. details: %w", def.Job.ID, def.Job.Type, err))

			err = jobModels.New().MarkTerminated(def.Job.ID, err.Error())
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

				err = jobModels.New().MarkTerminated(def.Job.ID, err.Error())
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
				err = jobModels.New().MarkTerminated(def.Job.ID, err.Error())
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("error marking job %s (%s) as terminated. details: %w", def.Job.ID, def.Job.Type, err))
					return
				}
				return
			} else {
				delay := int(math.Pow(2, float64(attempts-1)) * 30)
				runAfter := def.Job.RunAfter.Add(time.Duration(delay) * time.Second)
				utils.PrettyPrintError(fmt.Errorf("failed job %s (%s), attempt: %d delay: %ds details: %w", def.Job.ID, def.Job.Type, attempts-1, delay, err))

				err = jobModels.New().MarkFailed(def.Job.ID, runAfter, attempts, err.Error())
				if err != nil {
					utils.PrettyPrintError(fmt.Errorf("error marking job %s (%s) as failed. details: %w", def.Job.ID, def.Job.Type, err))
					return
				}
			}
		} else if strings.HasPrefix(err.Error(), "terminated") {
			err = errors.Unwrap(err)
			utils.PrettyPrintError(fmt.Errorf("terminated job %s (%s). details: %w", def.Job.ID, def.Job.Type, err))

			err = jobModels.New().MarkTerminated(def.Job.ID, err.Error())
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("error marking job %s (%s) as terminated. details: %w", def.Job.ID, def.Job.Type, err))
				return
			}
		} else {
			utils.PrettyPrintError(fmt.Errorf("error executing job %s (%s), terminating. details: %w", def.Job.ID, def.Job.Type, err))

			err = jobModels.New().MarkTerminated(def.Job.ID, err.Error())
			if err != nil {
				utils.PrettyPrintError(fmt.Errorf("error marking job %s (%s) as terminated. details: %w", def.Job.ID, def.Job.Type, err))
				return
			}
		}
	} else {
		err = jobModels.New().MarkCompleted(def.Job.ID)
		if err != nil {
			utils.PrettyPrintError(fmt.Errorf("error marking job %s (%s) as completed. details: %w", def.Job.ID, def.Job.Type, err))
			return
		}
	}
}
