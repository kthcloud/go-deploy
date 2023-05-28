package job_service

import (
	"fmt"
	jobModel "go-deploy/models/sys/job"
)

func Create(id, userID, jobType string, args map[string]interface{}) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create job. details: %s", err)
	}

	err := jobModel.CreateJob(id, userID, jobType, args)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func GetByID(userID, jobID string, isAdmin bool) (*jobModel.Job, error) {
	job, err := jobModel.GetByID(jobID)
	if err != nil {
		return nil, err
	}

	if job != nil && job.UserID != userID && !isAdmin {
		return nil, nil
	}

	return job, nil
}
