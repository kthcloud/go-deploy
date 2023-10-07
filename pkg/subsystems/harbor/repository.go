package harbor

import (
	"context"
	"fmt"
	"github.com/mittwald/goharbor-client/v5/apiv2/pkg/clients/artifact"
	"go-deploy/pkg/subsystems/harbor/models"
	"log"
	"strings"
)

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
		// for some reason, it is returned as name=<project name>/<repo name>
		// even though it is used as only <repo name> in the api
		// lovely harbor api :)
		repository.Name = name

		project, err := client.HarborClient.GetProject(context.TODO(), client.Project)
		if err != nil {
			return nil, makeError(err)
		}

		public = models.CreateRepositoryPublicFromGet(repository, project)
	}

	return public, nil
}

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
		return models.CreateRepositoryPublicFromGet(repository, nil), nil
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

	return models.CreateRepositoryPublicFromGet(repository, nil), nil
}

func (client *Client) UpdateRepository(public *models.RepositoryPublic) (*models.RepositoryPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update harbor repository %s. details: %w", public.Name, err)
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

	return nil, nil
}

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
