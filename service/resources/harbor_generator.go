package resources

import (
	"fmt"
	"go-deploy/pkg/config"
	"go-deploy/pkg/subsystems/harbor/models"
	"strings"
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
		Name:    hg.d.deployment.Name,
		Disable: false,
	}
}

func (hg *HarborGenerator) Repository() *models.RepositoryPublic {
	splits := strings.Split(config.Config.Registry.PlaceholderImage, "/")
	project := splits[len(splits)-2]
	repository := splits[len(splits)-1]

	return &models.RepositoryPublic{
		Name:   hg.d.deployment.Name,
		Seeded: false,
		Placeholder: &models.PlaceHolder{
			ProjectName:    project,
			RepositoryName: repository,
		},
	}
}

func (hg *HarborGenerator) Webhook() *models.WebhookPublic {
	webhookTarget := fmt.Sprintf("%s/v1/hooks/deployments/harbor", config.Config.ExternalUrl)
	return &models.WebhookPublic{
		Name:   hg.d.deployment.OwnerID,
		Target: webhookTarget,
		Token:  config.Config.Harbor.WebhookSecret,
	}
}
