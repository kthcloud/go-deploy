package resources

import (
	"fmt"
	"github.com/google/uuid"
	configModels "go-deploy/models/config"
	"go-deploy/models/model"
	"go-deploy/pkg/config"
	"go-deploy/pkg/subsystems"
	"go-deploy/pkg/subsystems/harbor/models"
	"go-deploy/service/generators"
	"strings"
)

type HarborGenerator struct {
	generators.HarborGeneratorBase

	deployment *model.Deployment
	zone       *configModels.Zone

	project string
}

func Harbor(deployment *model.Deployment, zone *configModels.Zone, project string) generators.HarborGenerator {
	return &HarborGenerator{
		deployment: deployment,
		zone:       zone,
		project:    project,
	}
}

// Project returns a models.ProjectPublic that should be created
func (hg *HarborGenerator) Project() *models.ProjectPublic {
	pr := models.ProjectPublic{
		Name:   hg.project,
		Public: false,
	}

	if p := &hg.deployment.Subsystems.Harbor.Project; subsystems.Created(p) {
		pr.ID = p.ID
		pr.CreatedAt = p.CreatedAt
	}

	return &pr
}

// Robot returns a models.RobotPublic that should be created
func (hg *HarborGenerator) Robot() *models.RobotPublic {
	if hg.deployment != nil {
		ro := models.RobotPublic{
			Name:    hg.deployment.Name,
			Disable: false,
		}

		if r := &hg.deployment.Subsystems.Harbor.Robot; subsystems.Created(r) {
			ro.ID = r.ID
			ro.HarborName = r.HarborName
			ro.Secret = r.Secret
			ro.CreatedAt = r.CreatedAt
		}

		return &ro
	}

	return nil
}

// Repository returns a models.RepositoryPublic that should be created
func (hg *HarborGenerator) Repository() *models.RepositoryPublic {
	splits := strings.Split(config.Config.Registry.PlaceholderImage, "/")
	project := splits[len(splits)-2]
	repository := splits[len(splits)-1]

	re := models.RepositoryPublic{
		Name: hg.deployment.Name,
		Placeholder: &models.PlaceHolder{
			ProjectName:    project,
			RepositoryName: repository,
		},
	}

	if r := &hg.deployment.Subsystems.Harbor.Repository; subsystems.Created(r) {
		re.ID = r.ID
		re.Seeded = r.Seeded
		re.CreatedAt = r.CreatedAt
	}

	return &re
}

// Webhook returns a models.WebhookPublic that should be created
func (hg *HarborGenerator) Webhook() *models.WebhookPublic {
	if hg.deployment != nil {
		webhookTarget := fmt.Sprintf("%s/v2/hooks/deployments/harbor", config.Config.ExternalUrl)

		we := models.WebhookPublic{
			// "Name" does not matter and will be imported from Harbor if "Target" matches with existing webhook
			Name:   uuid.NewString(),
			Target: webhookTarget,
			Token:  config.Config.Harbor.WebhookSecret,
		}

		if w := &hg.deployment.Subsystems.Harbor.Webhook; subsystems.Created(w) {
			we.ID = w.ID
			we.CreatedAt = w.CreatedAt
		}

		return &we
	}

	return nil
}
