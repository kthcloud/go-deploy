package harbor

import (
	"context"
	"fmt"
	"github.com/mittwald/goharbor-client/v5/apiv2/pkg/clients/artifact"
	"go-deploy/pkg/subsystems/harbor/models"
	"strings"
)

func (client *Client) insertPlaceholder(projectName, repositoryName string, placeholder *models.PlaceHolder) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to insert placeholder repository %s into %s for project %s. details: %s", placeholder, repositoryName, projectName, err)
	}

	fromArtifact, err := client.HarborClient.GetArtifact(context.TODO(), placeholder.ProjectName, placeholder.Repository, "latest")
	if err != nil {
		return makeError(err)
	}

	copyRef := &artifact.CopyReference{
		ProjectName:    placeholder.ProjectName,
		RepositoryName: placeholder.Repository,
		Tag:            "latest",
		Digest:         fromArtifact.Digest,
	}

	err = client.HarborClient.CopyArtifact(context.TODO(), copyRef, projectName, repositoryName)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func (client *Client) RepositoryCreated(projectName, name string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if harbor repository %s is created. details: %s", name, err)
	}

	repository, err := client.HarborClient.GetRepository(context.TODO(), projectName, name)
	if err != nil {
		return false, makeError(err)
	}

	return repository != nil, nil
}

func (client *Client) CreateRepository(projectName, name string, placeholder *models.PlaceHolder) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create harbor repository for deployment %s. details: %s", name, err)
	}

	projectArtifact, err := client.HarborClient.GetArtifact(context.TODO(), projectName, name, "latest")
	if err != nil {
		errStr := fmt.Sprintf("%s", err)
		is404 := strings.Contains(errStr, "getArtifactNotFound")
		if !is404 {
			return makeError(err)
		}
	}

	if projectArtifact != nil {
		return nil
	}

	if placeholder != nil {
		err = client.insertPlaceholder(projectName, name, placeholder)
		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

func (client *Client) DeleteRepository(projectName, name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete harbor repository %s. details: %s", name, err)
	}

	exists, project, err := client.assertProjectExists(projectName)
	if !exists {
		return nil
	}

	err = client.HarborClient.DeleteRepository(context.TODO(), project.Name, name)
	if err != nil {
		errStr := fmt.Sprintf("%s", err)
		if !strings.Contains(errStr, "deleteRepositoryNotFound") {
			return makeError(err)
		}
	}

	return nil
}
