package job_execute

import (
	"go-deploy/pkg/db/resources/job_repo"
	"go-deploy/pkg/jobs"
)

// JobFetcher is a worker that fetches new jobs from the database and runs them.
func JobFetcher() error {
	job, err := job_repo.New().GetNext()
	if err != nil {
		return err
	}

	if job == nil {
		return err
	}

	jobs.NewRunner(job).Run()

	return nil
}

// FailedJobFetcher is a worker that fetches failed jobs from the database and runs them.
func FailedJobFetcher() error {
	job, err := job_repo.New().GetNextFailed()
	if err != nil {
		return err
	}

	if job == nil {
		return err
	}

	jobs.NewRunner(job).Run()

	return nil
}
