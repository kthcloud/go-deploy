package job_execute

import (
	"context"

	"github.com/kthcloud/go-deploy/pkg/db/resources/job_repo"
	"github.com/kthcloud/go-deploy/pkg/jobs"
)

// JobFetcher is a worker that fetches new jobs from the database and runs them.
func JobFetcher(ctx context.Context) error {
	repo := job_repo.New()
	repo.Ctx = ctx
	job, err := repo.GetNext()
	if err != nil {
		return err
	}

	if job == nil {
		return nil
	}

	jobs.NewRunner(job).Run(ctx)

	return nil
}

// FailedJobFetcher is a worker that fetches failed jobs from the database and runs them.
func FailedJobFetcher(ctx context.Context) error {
	repo := job_repo.New()
	repo.Ctx = ctx
	job, err := repo.GetNextFailed()
	if err != nil {
		return err
	}

	if job == nil {
		return nil
	}

	jobs.NewRunner(job).Run(ctx)

	return nil
}
