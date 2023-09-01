package gitlab

import (
	"bytes"
	"fmt"
	"github.com/xanzy/go-gitlab"
	"go-deploy/pkg/subsystems/gitlab/models"
)

func (client *Client) ReadLastJob(projectID int) (*models.JobPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read last job. details: %w", err)
	}

	jobs, _, err := client.GitLabClient.Jobs.ListProjectJobs(projectID, &gitlab.ListJobsOptions{})
	if err != nil {
		return nil, makeError(err)
	}

	if len(jobs) > 0 {
		return models.CreateJobPublicFromGet(jobs[0]), nil
	}

	return nil, nil
}

func (client *Client) GetJobTrace(projectID int, jobID int) (string, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to get job trace. details: %w", err)
	}

	reader, _, err := client.GitLabClient.Jobs.GetTraceFile(projectID, jobID, nil)
	if err != nil {
		return "", makeError(err)
	}

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(reader)
	if err != nil {
		return "", makeError(err)
	}

	return buf.String(), nil
}
