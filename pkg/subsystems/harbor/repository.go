package harbor

import (
	"context"
	"fmt"
	artifactModels "github.com/kthcloud/go-deploy/pkg/imp/harbor/sdk/v2.0/client/artifact"
	repositoryModels "github.com/kthcloud/go-deploy/pkg/imp/harbor/sdk/v2.0/client/repository"
	"github.com/kthcloud/go-deploy/pkg/log"
	"github.com/kthcloud/go-deploy/pkg/subsystems/harbor/models"
)

// insertPlaceholder inserts a placeholder artifact into a repository.
// This is used to initialize a repository.
func (client *Client) insertPlaceholder(public *models.RepositoryPublic) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to insert placeholder repository %d. details: %w", public.ID, err)
	}

	_, err := client.HarborClient.V2().Artifact.CopyArtifact(context.TODO(), &artifactModels.CopyArtifactParams{
		From:           fmt.Sprintf("%s/%s:latest", public.Placeholder.ProjectName, public.Placeholder.RepositoryName),
		ProjectName:    client.Project,
		RepositoryName: public.Name,
	})
	if err != nil {
		return makeError(err)
	}

	return err
}

// ReadRepository reads a repository from Harbor.
func (client *Client) ReadRepository(name string) (*models.RepositoryPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read harbor repository %s. details: %w", name, err)
	}

	if name == "" {
		log.Println("Name not supplied when reading harbor repository. Assuming it was deleted")
		return nil, nil
	}

	repository, err := client.HarborClient.V2().Repository.GetRepository(context.TODO(), &repositoryModels.GetRepositoryParams{
		ProjectName:    client.Project,
		RepositoryName: name,
	})
	if err != nil {
		if IsNotFoundErr(err) {
			return nil, nil
		}

		return nil, makeError(err)
	}

	var public *models.RepositoryPublic
	if repository != nil {
		public = models.CreateRepositoryPublicFromGet(repository.Payload)
	}

	return public, nil
}

// CreateRepository creates a repository in Harbor.
func (client *Client) CreateRepository(public *models.RepositoryPublic) (*models.RepositoryPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create harbor repository %s. details: %w", public.Name, err)
	}

	repository, err := client.HarborClient.V2().Repository.GetRepository(context.TODO(), &repositoryModels.GetRepositoryParams{
		ProjectName:    client.Project,
		RepositoryName: public.Name,
	})
	if err != nil {
		if !IsNotFoundErr(err) {
			return nil, makeError(err)
		}
	}

	if repository != nil {
		return models.CreateRepositoryPublicFromGet(repository.Payload), nil
	}

	if public.Placeholder != nil {
		err = client.insertPlaceholder(public)
		if err != nil {
			return nil, makeError(err)
		}
	}

	return client.ReadRepository(public.Name)
}

// UpdateRepository updates a repository in Harbor.
func (client *Client) UpdateRepository(public *models.RepositoryPublic) (*models.RepositoryPublic, error) {
	// This does not actually update the repository, but the code is kept in case any field could be updated in the future
	return public, nil

	//makeError := func(err error) error {
	//	return fmt.Errorf("failed to update harbor repository %d. details: %w", public.ID, err)
	//}

	//if public.ID == 0 {
	//	log.Println("ID not supplied when updating harbor repository. Assuming it was deleted")
	//	return nil, nil
	//}
	//
	//repository, err := client.HarborClient.V2().Repository.GetRepository(context.Background(), &repositoryModels.GetRepositoryParams{
	//	ProjectName:    client.Project,
	//	RepositoryName: public.Name,
	//})
	//if err != nil {
	//	return nil, makeError(err)
	//}
	//
	//_, err = client.HarborClient.V2().Repository.UpdateRepository(context.Background(), &repositoryModels.UpdateRepositoryParams{
	//	ProjectName:    client.Project,
	//	Repository:     repository.Payload,
	//	RepositoryName: public.Name,
	//})
	//if err != nil {
	//	if IsNotFoundErr(err) {
	//		return nil, nil
	//	}
	//
	//	return nil, makeError(err)
	//}
	//
	//return client.ReadRepository(public.Name)
}

// DeleteRepository deletes a repository from Harbor.
func (client *Client) DeleteRepository(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete repository %s. details: %w", name, err)
	}

	if name == "" {
		log.Println("Name not supplied when deleting harbor repository. Assuming it was deleted")
		return nil
	}

	_, err := client.HarborClient.V2().Repository.DeleteRepository(context.TODO(), &repositoryModels.DeleteRepositoryParams{
		ProjectName:    client.Project,
		RepositoryName: name,
	})
	if err != nil {
		if IsNotFoundErr(err) {
			return nil
		}

		return makeError(err)
	}

	return nil
}
