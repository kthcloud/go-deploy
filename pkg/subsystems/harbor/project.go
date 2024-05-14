package harbor

import (
	"context"
	"fmt"
	projectModels "go-deploy/pkg/imp/harbor/sdk/v2.0/client/project"
	"go-deploy/pkg/log"
	"go-deploy/pkg/subsystems/harbor/models"
	"strconv"
)

// ReadProject reads a project from Harbor.
func (client *Client) ReadProject(id int) (*models.ProjectPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to read harbor project %d. details: %w", id, err)
	}

	if id == 0 {
		log.Println("ID not supplied when reading harbor project. Assuming it was deleted")
		return nil, nil
	}

	project, err := client.HarborClient.V2().Project.GetProject(context.TODO(), &projectModels.GetProjectParams{
		ProjectNameOrID: strconv.Itoa(id),
	})
	if err != nil {
		if IsNotFoundErr(err) {
			return nil, nil
		}

		return nil, makeError(err)
	}

	if project != nil {
		return models.CreateProjectPublicFromGet(project.Payload), nil
	}

	return nil, nil
}

// CreateProject creates a project in Harbor.
func (client *Client) CreateProject(public *models.ProjectPublic) (*models.ProjectPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to create harbor project %s. details: %w", public.Name, err)
	}

	project, err := client.HarborClient.V2().Project.GetProject(context.TODO(), &projectModels.GetProjectParams{
		ProjectNameOrID: public.Name,
	})
	if err != nil {
		if !IsNotFoundErr(err) {
			return nil, makeError(err)
		}
	}

	if project != nil {
		return models.CreateProjectPublicFromGet(project.Payload), nil
	}

	requestBody := models.CreateProjectCreateBody(public)
	_, err = client.HarborClient.V2().Project.CreateProject(context.TODO(), &projectModels.CreateProjectParams{
		Project: &requestBody,
	})
	if err != nil {
		return nil, makeError(err)
	}

	return client.ReadProject(public.ID)
}

// UpdateProject updates a project in Harbor.
func (client *Client) UpdateProject(public *models.ProjectPublic) (*models.ProjectPublic, error) {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update harbor project %d. details: %w", public.ID, err)
	}

	if public.ID == 0 {
		log.Println("ID not supplied when updating harbor project. Assuming it was deleted")
		return nil, nil
	}

	requestBody := models.CreateProjectUpdateParamsFromPublic(public)
	_, err := client.HarborClient.V2().Project.UpdateProject(context.TODO(), &projectModels.UpdateProjectParams{
		Project:         requestBody,
		ProjectNameOrID: strconv.Itoa(public.ID),
	})
	if err != nil {
		if IsNotFoundErr(err) {
			return nil, nil
		}

		return nil, makeError(err)
	}

	project, err := client.ReadProject(public.ID)
	if err != nil {
		return nil, makeError(err)
	}

	if project != nil {
		return project, nil
	}

	log.Println("Harbor project", public.Name, "not found when updating. Assuming it was deleted")
	return nil, nil
}

// DeleteProject deletes a project from Harbor.
func (client *Client) DeleteProject(id int) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to delete project %d. details: %w", id, err)
	}

	if id == 0 {
		log.Println("ID not supplied when deleting harbor project. Assuming it was deleted")
		return nil
	}

	_, err := client.HarborClient.V2().Project.DeleteProject(context.TODO(), &projectModels.DeleteProjectParams{
		ProjectNameOrID: strconv.Itoa(id),
	})
	if err != nil {
		if IsNotFoundErr(err) {
			return nil
		}

		return makeError(err)
	}

	return nil
}
