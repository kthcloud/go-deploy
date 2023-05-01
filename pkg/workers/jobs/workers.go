package jobs

import (
	jobModel "go-deploy/models/job"
	"go-deploy/pkg/app"
	"time"
)

func jobFetcher(ctx *app.Context) {
	for {
		if ctx.Stop {
			break
		}

		job, err := jobModel.GetNextJob()
		if err != nil {
			continue
		}

		if job == nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		switch job.Type {
		case "createDeployment":
			go createDeployment(job)

		}
	}
}
