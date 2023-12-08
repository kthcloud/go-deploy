package github

import (
	"context"
	"fmt"
	"go-deploy/pkg/subsystems/github/models"
	"log"
)

func (client *Client) ReadRepository(id int64) (*models.RepositoryPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read repository %d. details: %w", id, err)
	}

	repo, _, err := client.GitHubClient.Repositories.GetByID(context.Background(), id)
	if err != nil {
		return nil, makeError(err)
	}

	if repo.ID == nil || repo.Name == nil || repo.CloneURL == nil || repo.DefaultBranch == nil {
		log.Println("failed to get github repository. one of the fields is nil. assuming it was deleted")
		return nil, nil
	}

	return models.CreateRepositoryPublicFromRead(repo), nil
}
