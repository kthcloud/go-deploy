package harbor

import (
	"context"
	"fmt"
	"github.com/mittwald/goharbor-client/v5/apiv2/pkg/clients/artifact"
	"go-deploy/pkg/subsystems/harbor/models"
	"strings"
)

func (client *Client) insertPlaceholder(public *models.RepositoryPublic) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to insert placeholder repository %d. details: %s", public.ID, err)
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

	err = client.HarborClient.CopyArtifact(context.TODO(), copyRef, public.ProjectName, public.Name)
	if err != nil {
		return makeError(err)
	}

	return err
}

func (client *Client) RepositoryCreated(public *models.RepositoryPublic) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if repository %s is created. details: %s", public.Name, err)
	}

	repository, err := client.HarborClient.GetRepository(context.TODO(), public.ProjectName, public.Name)
	if err != nil {
		return false, makeError(err)
	}

	return repository != nil, nil
}

func (client *Client) ReadRepository(projectName, name string) (*models.RepositoryPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read repository %s. details: %s", name, err)
	}

	repository, err := client.HarborClient.GetRepository(context.TODO(), projectName, name)
	if err != nil {
		errStr := fmt.Sprintf("%s", err)
		if !strings.Contains(errStr, "getRepositoryNotFound") {
			return nil, makeError(err)
		}
	}
	// for some reason it is returned as name=<project name>/<repo name>
	// even though it is used as only <repo name> in the api
	repository.Name = name

	project, err := client.HarborClient.GetProject(context.TODO(), projectName)

	public := models.CreateRepositoryPublicFromGet(repository, project)

	return public, nil
}

func (client *Client) CreateRepository(public *models.RepositoryPublic) (int, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create repository %s. details: %s", public.Name, err)
	}

	repository, err := client.HarborClient.GetRepository(context.TODO(), public.ProjectName, public.Name)
	if err != nil {
		errStr := fmt.Sprintf("%s", err)
		if !strings.Contains(errStr, "getRepositoryNotFound") {
			return 0, makeError(err)
		}
	}

	if repository != nil {
		return int(repository.ID), nil
	}

	if public.Placeholder != nil {
		err = client.insertPlaceholder(public)
		if err != nil {
			return 0, makeError(err)
		}
	}

	repository, err = client.HarborClient.GetRepository(context.TODO(), public.ProjectName, public.Name)
	if err != nil {
		return 0, makeError(err)
	}

	return int(repository.ID), nil
}

func (client *Client) DeleteRepository(projectName, name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete repository %s. details: %s", name, err)
	}

	err := client.HarborClient.DeleteRepository(context.TODO(), projectName, name)
	if err != nil {
		errStr := fmt.Sprintf("%s", err)
		if !strings.Contains(errStr, "deleteRepositoryNotFound") {
			return makeError(err)
		}
	}

	return nil
}
