package harbor

import (
	"context"
	"fmt"
	"github.com/mittwald/goharbor-client/v5/apiv2/pkg/clients/artifact"
	"go-deploy/pkg/subsystems/harbor/models"
	"log"
	"strings"
)

// insertPlaceholder inserts a placeholder artifact into a repository.
// This is used to initialize a repository.
func (client *Client) insertPlaceholder(public *models.RepositoryPublic) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to insert placeholder repository %d. details: %w", public.ID, err)
	}

	fromArtifact, err := client.HarborClient.GetArtifact(context.TODO(), public.Placeholder.ProjectName, public.Placeholder.RepositoryName, "latest")
	if err != nil {
		return makeError(err)
	}

	copyRef := &artifact.CopyReference{
		ProjectName:    public.Placeholder.ProjectName,
		RepositoryName: public.Placeholder.RepositoryName,
		Tag:            "latest",
		Digest:         fromArtifact.Digest,
	}

	err = client.HarborClient.CopyArtifact(context.TODO(), copyRef, client.Project, public.Name)
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
		log.Println("name not supplied when reading harbor repository. assuming it was deleted")
		return nil, nil
	}

	repository, err := client.HarborClient.GetRepository(context.TODO(), client.Project, name)
	if err != nil {
		errStr := fmt.Sprintf("%s", err)
		if !strings.Contains(errStr, "NotFound") {
			return nil, makeError(err)
		}
	}

	var public *models.RepositoryPublic
	if repository != nil {
		public = models.CreateRepositoryPublicFromGet(repository)
	}

	return public, nil
}

// CreateRepository creates a repository in Harbor.
func (client *Client) CreateRepository(public *models.RepositoryPublic) (*models.RepositoryPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create harbor repository %s. details: %w", public.Name, err)
	}

	repository, err := client.HarborClient.GetRepository(context.TODO(), client.Project, public.Name)
	if err != nil {
		errStr := fmt.Sprintf("%s", err)
		if !strings.Contains(errStr, "getRepositoryNotFound") {
			return nil, makeError(err)
		}
	}

	if repository != nil {
		return models.CreateRepositoryPublicFromGet(repository), nil
	}

	if public.Placeholder != nil {
		err = client.insertPlaceholder(public)
		if err != nil {
			return nil, makeError(err)
		}
	}

	repository, err = client.HarborClient.GetRepository(context.TODO(), client.Project, public.Name)
	if err != nil {
		return nil, makeError(err)
	}

	return models.CreateRepositoryPublicFromGet(repository), nil
}

// UpdateRepository updates a repository in Harbor.
func (client *Client) UpdateRepository(public *models.RepositoryPublic) (*models.RepositoryPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update harbor repository %d. details: %w", public.ID, err)
	}

	if public.ID == 0 {
		log.Println("id not supplied when updating harbor repository. assuming it was deleted")
		return nil, nil
	}

	repository, err := client.HarborClient.GetRepository(context.TODO(), client.Project, public.Name)
	if err != nil {
		return nil, makeError(err)
	}

	if repository == nil {
		return nil, makeError(fmt.Errorf("repository %s not found", public.Name))
	}

	// this doesn't actually do anything, but the code here is kept in case any field could be updated in the future

	err = client.HarborClient.UpdateRepository(context.TODO(), client.Project, public.Name, repository)
	if err != nil {
		return nil, makeError(err)
	}

	return client.ReadRepository(public.Name)
}

// DeleteRepository deletes a repository from Harbor.
func (client *Client) DeleteRepository(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete repository %s. details: %w", name, err)
	}

	if name == "" {
		log.Println("name not supplied when deleting harbor repository. assuming it was deleted")
		return nil
	}

	err := client.HarborClient.DeleteRepository(context.TODO(), client.Project, name)
	if err != nil {
		errStr := fmt.Sprintf("%s", err)
		if !strings.Contains(errStr, "deleteRepositoryNotFound") {
			return makeError(err)
		}
	}

	return nil
}
