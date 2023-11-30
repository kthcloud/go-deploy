package resources

import (
	"fmt"
	"go-deploy/pkg/config"
	"go-deploy/pkg/subsystems/harbor/models"
	"go-deploy/service"
	"strings"
)

type HarborGenerator struct {
	*PublicGeneratorType
	project string
}

func (hg *HarborGenerator) Project() *models.ProjectPublic {
	if hg.d.deployment != nil {
		pr := models.ProjectPublic{
			Name:   hg.project,
			Public: false,
		}

		if p := &hg.d.deployment.Subsystems.Harbor.Project; service.Created(p) {
			pr.ID = p.ID
			pr.CreatedAt = p.CreatedAt
		}

		return &pr
	}

	return nil
}

func (hg *HarborGenerator) Robot() *models.RobotPublic {
	if hg.d.deployment != nil {
		ro := models.RobotPublic{
			Name:    hg.d.deployment.Name,
			Disable: false,
		}

		if r := &hg.d.deployment.Subsystems.Harbor.Robot; service.Created(r) {
			ro.ID = r.ID
			ro.HarborName = r.HarborName
			ro.Secret = r.Secret
			ro.CreatedAt = r.CreatedAt

		}

		return &ro
	}

	return nil
}

func (hg *HarborGenerator) Repository() *models.RepositoryPublic {
	if hg.d.deployment != nil {
		splits := strings.Split(config.Config.Registry.PlaceholderImage, "/")
		project := splits[len(splits)-2]
		repository := splits[len(splits)-1]

		re := models.RepositoryPublic{
			Name: hg.d.deployment.Name,
			Placeholder: &models.PlaceHolder{
				ProjectName:    project,
				RepositoryName: repository,
			},
		}

		if r := &hg.d.deployment.Subsystems.Harbor.Repository; service.Created(r) {
			re.ID = r.ID
			re.Seeded = r.Seeded
			re.CreatedAt = r.CreatedAt
		}

		return &re
	}

	return nil
}

func (hg *HarborGenerator) Webhook() *models.WebhookPublic {
	if hg.d.deployment != nil {
		webhookTarget := fmt.Sprintf("%s/v1/hooks/deployments/harbor", config.Config.ExternalUrl)

		we := models.WebhookPublic{
			Name:   hg.d.deployment.OwnerID,
			Target: webhookTarget,
			Token:  config.Config.Harbor.WebhookSecret,
		}

		if w := &hg.d.deployment.Subsystems.Harbor.Webhook; service.Created(w) {
			we.ID = w.ID
			we.CreatedAt = w.CreatedAt
		}

		return &we
	}

	return nil
}
