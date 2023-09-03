package gitlab_service

import (
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/models/sys/deployment/subsystems"
	"go-deploy/pkg/subsystems/gitlab/models"
)

func updateGitLabBuild(deploymentID string, lastJob *models.JobPublic, trace []string) error {
	return deploymentModel.New().UpdateGitLabBuild(deploymentID, subsystems.GitLabBuild{
		ID:        lastJob.ID,
		ProjectID: lastJob.ProjectID,
		Trace:     trace,
		Status:    lastJob.Status,
		Stage:     lastJob.Stage,
		CreatedAt: lastJob.CreatedAt,
	})
}
