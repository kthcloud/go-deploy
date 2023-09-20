package helpers

import (
	"fmt"
	"github.com/google/uuid"
	"go-deploy/pkg/conf"
	harborModels "go-deploy/pkg/subsystems/harbor/models"
)

func CreateProjectPublic(projectName string) *harborModels.ProjectPublic {
	return &harborModels.ProjectPublic{
		Name: projectName,
	}
}

func CreateRobotPublic(name string, projectID int, projectName string) *harborModels.RobotPublic {
	return &harborModels.RobotPublic{
		Name:        name,
		ProjectID:   projectID,
		ProjectName: projectName,
		Disable:     false,
	}
}

func CreateRepositoryPublic(projectID int, projectName string, name string) *harborModels.RepositoryPublic {
	return &harborModels.RepositoryPublic{
		Name:        name,
		ProjectID:   projectID,
		ProjectName: projectName,
		Seeded:      false,
		Placeholder: &harborModels.PlaceHolder{
			ProjectName:    conf.Env.DockerRegistry.Placeholder.Project,
			RepositoryName: conf.Env.DockerRegistry.Placeholder.Repository,
		},
	}
}

func CreateWebhookPublic(projectID int, projectName string) *harborModels.WebhookPublic {
	webhookTarget := fmt.Sprintf("%s/v1/hooks/deployments/harbor", conf.Env.ExternalUrl)
	return &harborModels.WebhookPublic{
		Name:        uuid.NewString(),
		ProjectID:   projectID,
		ProjectName: projectName,
		Target:      webhookTarget,
		Token:       conf.Env.Harbor.WebhookSecret,
	}
}
