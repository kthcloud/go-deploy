package harbor_service

import (
	"errors"
	"fmt"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/service"
	"go-deploy/service/deployment_service/base"
	"go-deploy/service/deployment_service/resources"
	"log"
)

func Create(deploymentID, userID string, params *deploymentModel.CreateParams) error {
	log.Println("setting up harbor for", params.Name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup harbor for deployment %s. details: %w", params.Name, err)
	}

	context, err := NewContext(deploymentID)
	if err != nil {
		if errors.Is(err, base.DeploymentDeletedErr) {
			return nil
		}

		return makeError(err)
	}

	context.WithCreateParams(params)

	// Project
	err = resources.SsCreator(context.Client.CreateProject).
		WithID(context.Deployment.ID).
		WithPublic(context.Generator.Project()).
		WithDbKey("harbor.project").
		Exec()

	if err != nil {
		return makeError(err)
	}

	// Robot
	err = resources.SsCreator(context.Client.CreateRobot).
		WithID(context.Deployment.ID).
		WithPublic(context.Generator.Robot()).
		WithDbKey("harbor.robot").
		Exec()

	if err != nil {
		return makeError(err)
	}

	// Repository
	err = resources.SsCreator(context.Client.CreateRepository).
		WithID(context.Deployment.ID).
		WithPublic(context.Generator.Repository()).
		WithDbKey("harbor.repository").
		Exec()

	if err != nil {
		return makeError(err)
	}

	// Webhook
	err = resources.SsCreator(context.Client.CreateWebhook).
		WithID(context.Deployment.ID).
		WithPublic(context.Generator.Webhook()).
		WithDbKey("harbor.webhook").
		Exec()

	if err != nil {
		return makeError(err)
	}

	return nil
}

func CreatePlaceholder(id string) error {
	log.Println("setting up placeholder harbor")

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup placeholder harbor. details: %w", err)
	}

	context, err := NewContext(id)
	if err != nil {
		if errors.Is(err, base.DeploymentDeletedErr) {
			return nil
		}

		return makeError(err)
	}

	return resources.SsPlaceholderCreator().WithID(context.Deployment.ID).WithDbKey("harbor").Exec()
}

func Delete(id string) error {
	log.Println("deleting harbor for", id)

	makeError := func(err error) error {
		return fmt.Errorf("failed to delete harbor for deployment %s. details: %w", id, err)
	}

	context, err := NewContext(id)
	if err != nil {
		if errors.Is(err, base.DeploymentDeletedErr) {
			return nil
		}

		return makeError(err)
	}

	if hook := &context.Deployment.Subsystems.Harbor.Webhook; service.Created(hook) {
		err = resources.SsDeleter(context.Client.DeleteWebhook).
			WithID(context.Deployment.ID).
			WithResourceID(hook.ID).
			WithDbKey("harbor.webhook").
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	if repo := &context.Deployment.Subsystems.Harbor.Repository; service.Created(repo) {
		err = resources.SsDeleter(context.Client.DeleteRepository).
			WithID(context.Deployment.ID).
			WithResourceID(repo.Name).
			WithDbKey("harbor.repository").
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	if robot := &context.Deployment.Subsystems.Harbor.Robot; service.Created(robot) {
		err = resources.SsDeleter(context.Client.DeleteRobot).
			WithID(context.Deployment.ID).
			WithResourceID(robot.ID).
			WithDbKey("harbor.robot").
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	if project := &context.Deployment.Subsystems.Harbor.Project; service.Created(project) {
		err = resources.SsDeleter(func(int) error { return nil }).
			WithID(context.Deployment.ID).
			WithResourceID(project.ID).
			WithDbKey("harbor.project").
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

func Repair(name string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to repair harbor %s. details: %w", name, err)
	}

	context, err := NewContext(name)
	if err != nil {
		if errors.Is(err, base.DeploymentDeletedErr) {
			return nil
		}

		return makeError(err)
	}

	if project := &context.Deployment.Subsystems.Harbor.Project; service.Created(project) {
		err = resources.SsRepairer(
			context.Client.ReadProject,
			context.Client.CreateProject,
			context.Client.UpdateProject,
			func(int) error { return nil },
		).WithID(context.Deployment.ID).WithResourceID(project.ID).WithDbKey("harbor.project").Exec()

		if err != nil {
			return makeError(err)
		}
	}

	if robot := &context.Deployment.Subsystems.Harbor.Robot; service.Created(robot) {
		err = resources.SsRepairer(
			context.Client.ReadRobot,
			context.Client.CreateRobot,
			context.Client.UpdateRobot,
			context.Client.DeleteRobot,
		).WithID(context.Deployment.ID).WithResourceID(robot.ID).WithDbKey("harbor.robot").Exec()

		if err != nil {
			return makeError(err)
		}
	}

	if repo := &context.Deployment.Subsystems.Harbor.Repository; service.Created(repo) {
		err = resources.SsRepairer(
			context.Client.ReadRepository,
			context.Client.CreateRepository,
			context.Client.UpdateRepository,
			context.Client.DeleteRepository,
		).WithID(context.Deployment.ID).WithResourceID(repo.Name).WithDbKey("harbor.repository").Exec()

		if err != nil {
			return makeError(err)
		}
	}

	if hook := &context.Deployment.Subsystems.Harbor.Webhook; service.Created(hook) {
		err = resources.SsRepairer(
			context.Client.ReadWebhook,
			context.Client.CreateWebhook,
			context.Client.UpdateWebhook,
			context.Client.DeleteWebhook,
		).WithID(context.Deployment.ID).WithResourceID(hook.ID).WithDbKey("harbor.webhook").Exec()

		if err != nil {
			return makeError(err)
		}
	}

	return nil
}
