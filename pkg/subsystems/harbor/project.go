package harbor

import (
	"context"
	"errors"
	"fmt"
	"github.com/mittwald/goharbor-client/v5/apiv2/pkg/common"
	harborErrors "github.com/mittwald/goharbor-client/v5/apiv2/pkg/errors"
	"go-deploy/pkg/subsystems/harbor/models"
	"strconv"
	"strings"
)

func (client *Client) ProjectCreated(id int) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if project %d is created. details: %s", id, err)
	}

	project, err := client.HarborClient.GetProject(context.TODO(), strconv.Itoa(id))
	if err != nil {
		targetErr := &harborErrors.ErrProjectNotFound{}
		if errors.As(err, &targetErr) {
			return false, nil
		} else {
			return false, makeError(err)
		}
	}

	if project == nil || project.ProjectID == 0 || project.Metadata.Public == "false" {
		return false, nil
	}

	return true, nil
}

func (client *Client) ProjectDeleted(id int) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if project %d is created. details: %s", id, err)
	}

	_, err := client.HarborClient.GetProject(context.TODO(), strconv.Itoa(id))
	if err != nil {
		targetErr := &harborErrors.ErrProjectNotFound{}
		if !errors.As(err, &targetErr) {
			return false, makeError(err)
		}
	}

	return true, nil
}

func (client *Client) ReadProject(id int) (*models.ProjectPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read project %d. details: %s", id, err.Error())
	}

	project, err := client.HarborClient.GetProject(context.TODO(), strconv.Itoa(id))
	if err != nil {
		targetErr := &harborErrors.ErrProjectNotFound{}
		if !errors.As(err, &targetErr) {
			return nil, makeError(err)
		}
	}

	var public *models.ProjectPublic
	if project != nil {
		public = models.CreateProjectPublicFromGet(project)
	}

	return public, nil
}

func (client *Client) CreateProject(public *models.ProjectPublic) (int, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create project %s. details: %s", public.Name, err.Error())
	}

	if public.Name == "" {
		return 0, makeError(errors.New("project name is empty"))
	}

	project, err := client.HarborClient.GetProject(context.TODO(), public.Name)
	if err != nil {
		targetErr := &harborErrors.ErrProjectNotFound{}
		if !errors.As(err, &targetErr) {
			return 0, makeError(err)
		}
	}

	if project != nil {
		return int(project.ProjectID), nil
	}

	requestBody := models.CreateProjectCreateBody(public)
	err = client.HarborClient.NewProject(context.TODO(), &requestBody)
	if err != nil {
		return 0, makeError(err)
	}

	project, err = client.HarborClient.GetProject(context.TODO(), public.Name)
	if err != nil {
		return 0, makeError(err)
	}

	return int(project.ProjectID), nil
}

func (client *Client) DeleteProject(id int) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete project %d. details: %s", id, err)
	}

	err := client.HarborClient.DeleteProject(context.TODO(), strconv.Itoa(id))
	if err != nil {
		errString := fmt.Sprintf("%s", err)
		if !strings.Contains(errString, "not found on server side") {
			return makeError(err)
		}
	}

	return nil
}

func (client *Client) UpdateProject(public *models.ProjectPublic) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update project %s. details: %s", public.Name, err)
	}

	requestBody := models.CreateProjectUpdateParamsFromPublic(public)
	err := client.HarborClient.UpdateProject(context.TODO(), requestBody, nil)
	if err != nil {
		errString := fmt.Sprintf("%s", err)
		if !strings.Contains(errString, "id/name pair not found on server side") {
			return makeError(err)
		}
	}

	err = client.HarborClient.UpdateProjectMetadata(
		context.TODO(),
		public.Name,
		common.ProjectMetadataKeyPublic,
		boolToString(public.Public),
	)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func (client *Client) IsProjectEmpty(id int) (bool, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to check if project %d is empty. details: %s", id, err)
	}

	project, err := client.HarborClient.GetProject(context.TODO(), strconv.Itoa(id))
	if err != nil {
		errString := fmt.Sprintf("%s", err)
		if !strings.Contains(errString, "project not found on server side") {
			return false, makeError(err)
		}
	}

	if project == nil {
		return true, nil
	}

	repositories, err := client.HarborClient.ListRepositories(context.TODO(), project.Name)
	if err != nil {
		return false, makeError(err)
	}

	return len(repositories) == 0, nil
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
