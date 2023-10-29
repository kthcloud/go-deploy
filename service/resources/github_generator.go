package resources

import (
	"fmt"
	"github.com/google/uuid"
	"go-deploy/pkg/conf"
	githubModels "go-deploy/pkg/subsystems/github/models"
)

type GitHubGenerator struct {
	*PublicGeneratorType
	token        string
	repositoryID int64
}

func (gg *GitHubGenerator) Webhook() *githubModels.WebhookPublic {
	if gg.d.deployment != nil {
		webhookTarget := fmt.Sprintf("%s/v1/hooks/deployments/github", conf.Env.ExternalUrl)
		return &githubModels.WebhookPublic{
			RepositoryID: gg.repositoryID,
			Events:       nil,
			Active:       false,
			ContentType:  "json",
			WebhookURL:   webhookTarget,
			Secret:       uuid.NewString(),
		}
	}

	return nil
}
