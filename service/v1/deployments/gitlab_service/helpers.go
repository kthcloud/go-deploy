package gitlab_service

import (
	deploymentModels "go-deploy/models/sys/deployment"
	"go-deploy/models/sys/deployment/subsystems"
	"go-deploy/pkg/subsystems/gitlab/models"
)

// updateGitLabBuild updates the GitLab build for the deployment.
//
// It updates the GitLab build with the last job, and the trace.
func updateGitLabBuild(deploymentID string, lastJob *models.JobPublic, trace []string) error {
	return deploymentModels.New().UpdateGitLabBuild(deploymentID, subsystems.GitLabBuild{
		ID:        lastJob.ID,
		ProjectID: lastJob.ProjectID,
		Trace:     trace,
		Status:    lastJob.Status,
		Stage:     lastJob.Stage,
		CreatedAt: lastJob.CreatedAt,
	})
}
