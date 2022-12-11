package harbor

import (
	"context"
	"errors"
	"fmt"
	harborErrors "github.com/mittwald/goharbor-client/v5/apiv2/pkg/errors"
	"go-deploy/pkg/subsystems/harbor/models"
	"strings"
)

func (client *Client) ProjectCreated(name string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if harbor project %s is created. details: %s", name, err)
	}

	project, err := client.HarborClient.GetProject(context.TODO(), name)
	if err != nil {
		return false, makeError(err)
	}

	if project == nil {
		return false, nil
	}

	publicProject := project.Metadata.Public == "true"

	return publicProject, nil
}

func (client *Client) ProjectDeleted(name string) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if harbor project %s is created. details: %s", name, err)
	}

	_, err := client.HarborClient.GetProject(context.TODO(), name)
	if err != nil {
		targetErr := &harborErrors.ErrProjectNotFound{}
		if !errors.As(err, &targetErr) {
			return false, makeError(err)
		}
	}

	return true, nil
}

func (client *Client) CreateProject(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create harbor project %s. details: %s", name, err.Error())
	}

	project, err := client.HarborClient.GetProject(context.TODO(), name)
	if err != nil {
		targetErr := &harborErrors.ErrProjectNotFound{}
		if !errors.As(err, &targetErr) {
			return makeError(err)
		}
	}

	if project == nil {
		requestBody := models.CreateProjectCreateReq(name)
		err = client.HarborClient.NewProject(context.TODO(), &requestBody)
		if err != nil {
			return makeError(err)
		}

		err = client.HarborClient.UpdateProjectMetadata(context.TODO(), name, "public", "true")
		if err != nil {
			return makeError(err)
		}
	} else if project.Metadata.Public == "false" {
		err = client.HarborClient.UpdateProjectMetadata(context.TODO(), name, "public", "true")
		if err != nil {
			return makeError(err)
		}
	}
	return nil
}

func (client *Client) DeleteProject(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete harbor project %s. details: %s", name, err)
	}

	err := client.HarborClient.DeleteProject(context.TODO(), name)
	if err != nil {
		errString := fmt.Sprintf("%s", err)
		if strings.Contains(errString, "id/name pair not found on server side") {
			return nil
		}

		return makeError(err)
	}

	return nil
}
