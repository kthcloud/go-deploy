package resources

import (
	"fmt"
	"github.com/google/uuid"
	"go-deploy/pkg/config"
	"go-deploy/pkg/subsystems"
	githubModels "go-deploy/pkg/subsystems/github/models"
)

type GitHubGenerator struct {
	*PublicGeneratorType
	token        string
	repositoryID int64
}

func (gg *GitHubGenerator) Webhook() *githubModels.WebhookPublic {
	if gg.d.deployment != nil {
		webhookTarget := fmt.Sprintf("%s/v1/hooks/deployments/github", config.Config.ExternalUrl)
		wh := githubModels.WebhookPublic{
			RepositoryID: gg.repositoryID,
			Events:       nil,
			Active:       false,
			ContentType:  "json",
			WebhookURL:   webhookTarget,
			Secret:       uuid.NewString(),
		}

		if w := &gg.d.deployment.Subsystems.GitHub.Webhook; subsystems.Created(w) {
			wh.ID = w.ID
			wh.CreatedAt = w.CreatedAt
		}

		return &wh
	}

	return nil
}
