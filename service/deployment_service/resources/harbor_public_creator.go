package resources

import (
	"fmt"
	"go-deploy/pkg/conf"
	"go-deploy/pkg/subsystems/harbor/models"
)

type HarborGenerator struct {
	*PublicGeneratorType
	project string
}

func (hg *HarborGenerator) Project() *models.ProjectPublic {
	return &models.ProjectPublic{
		Name:   hg.project,
		Public: false,
	}
}

func (hg *HarborGenerator) Robot() *models.RobotPublic {
	return &models.RobotPublic{
		Name:    hg.deployment.Name,
		Disable: false,
	}
}

func (hg *HarborGenerator) Repository() *models.RepositoryPublic {
	return &models.RepositoryPublic{
		Name:   hg.deployment.Name,
		Seeded: false,
		Placeholder: &models.PlaceHolder{
			ProjectName:    conf.Env.DockerRegistry.Placeholder.Project,
			RepositoryName: conf.Env.DockerRegistry.Placeholder.Repository,
		},
	}
}

func (hg *HarborGenerator) Webhook() *models.WebhookPublic {
	webhookTarget := fmt.Sprintf("%s/v1/hooks/deployments/harbor", conf.Env.ExternalUrl)
	return &models.WebhookPublic{
		Name:   hg.deployment.OwnerID,
		Target: webhookTarget,
		Token:  conf.Env.Harbor.WebhookSecret,
	}
}
