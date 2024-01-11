package harbor

import (
	"context"
	"errors"
	"fmt"
	harborErrors "github.com/mittwald/goharbor-client/v5/apiv2/pkg/errors"
	"go-deploy/pkg/subsystems/harbor/models"
	"log"
	"strconv"
	"strings"
)

func (client *Client) ReadProject(id int) (*models.ProjectPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read harbor project %d. details: %w", id, err)
	}

	if id == 0 {
		log.Println("id not supplied when reading harbor project. assuming it was deleted")
		return nil, nil
	}

	project, err := client.HarborClient.GetProject(context.TODO(), strconv.Itoa(id))
	if err != nil {
		targetErr := &harborErrors.ErrProjectNotFound{}
		if !errors.As(err, &targetErr) {
			return nil, makeError(err)
		}
	}

	if project != nil {
		return models.CreateProjectPublicFromGet(project), nil
	}

	return nil, nil
}

func (client *Client) CreateProject(public *models.ProjectPublic) (*models.ProjectPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create harbor project %s. details: %w", public.Name, err)
	}

	project, err := client.HarborClient.GetProject(context.TODO(), public.Name)
	if err != nil {
		targetErr := &harborErrors.ErrProjectNotFound{}
		if !errors.As(err, &targetErr) {
			return nil, makeError(err)
		}
	}

	if project != nil {
		return models.CreateProjectPublicFromGet(project), nil
	}

	requestBody := models.CreateProjectCreateBody(public)
	err = client.HarborClient.NewProject(context.TODO(), &requestBody)
	if err != nil {
		return nil, makeError(err)
	}

	project, err = client.HarborClient.GetProject(context.TODO(), public.Name)
	if err != nil {
		return nil, makeError(err)
	}

	return models.CreateProjectPublicFromGet(project), nil
}

func (client *Client) UpdateProject(public *models.ProjectPublic) (*models.ProjectPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update harbor project %d. details: %w", public.ID, err)
	}

	if public.ID == 0 {
		log.Println("id not supplied when updating harbor project. assuming it was deleted")
		return nil, nil
	}

	requestBody := models.CreateProjectUpdateParamsFromPublic(public)
	err := client.HarborClient.UpdateProject(context.TODO(), requestBody, nil)
	if err != nil {
		// use ErrProjectMismatchMsg
		mismatch := &harborErrors.ErrProjectMismatch{}
		notFound := &harborErrors.ErrProjectNotFound{}
		if !errors.As(err, &mismatch) && !errors.As(err, &notFound) {
			return nil, makeError(err)
		}
	}

	project, err := client.ReadProject(public.ID)
	if err != nil {
		mismatch := &harborErrors.ErrProjectMismatch{}
		notFound := &harborErrors.ErrProjectNotFound{}
		if !errors.As(err, &mismatch) && !errors.As(err, &notFound) {
			return nil, makeError(err)
		}

		return nil, makeError(err)
	}

	if project != nil {
		return project, nil
	}

	log.Println("harbor project", public.Name, "not found when updating. assuming it was deleted")
	return nil, nil
}

func (client *Client) DeleteProject(id int) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete project %d. details: %w", id, err)
	}

	if id == 0 {
		log.Println("id not supplied when deleting harbor project. assuming it was deleted")
		return nil
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
