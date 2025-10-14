package generators

import (
	"github.com/kthcloud/go-deploy/pkg/subsystems/harbor/models"
)

// HarborGenerator is a generator for Harbor resources
// It is used to generate the `publics`, such as models.ProjectPublic and models.RobotPublic
type HarborGenerator interface {
	Project() *models.ProjectPublic
	Robot() *models.RobotPublic
	Repository() *models.RepositoryPublic
	Webhook() *models.WebhookPublic
}

type HarborGeneratorBase struct {
	HarborGenerator
}

func (hg *HarborGeneratorBase) Project() *models.ProjectPublic {
	return nil
}

func (hg *HarborGeneratorBase) Robot() *models.RobotPublic {
	return nil
}

func (hg *HarborGeneratorBase) Repository() *models.RepositoryPublic {
	return nil
}

func (hg *HarborGeneratorBase) Webhook() *models.WebhookPublic {
	return nil
}
