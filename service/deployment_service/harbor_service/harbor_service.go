package harbor_service

import (
	"errors"
	"fmt"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/service"
	"go-deploy/service/deployment_service/base"
	"go-deploy/service/resources"
	"go-deploy/utils/subsystemutils"
	"log"
)

func Create(deploymentID string, params *deploymentModel.CreateParams) error {
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

	// Project
	err = resources.SsCreator(context.Client.CreateProject).
		WithDbFunc(dbFunc(deploymentID, "project")).
		WithPublic(context.Generator.Project()).
		Exec()

	if err != nil {
		return makeError(err)
	}

	// Robot
	err = resources.SsCreator(context.Client.CreateRobot).
		WithDbFunc(dbFunc(deploymentID, "robot")).
		WithPublic(context.Generator.Robot()).
		Exec()

	if err != nil {
		return makeError(err)
	}

	// Repository
	err = resources.SsCreator(context.Client.CreateRepository).
		WithDbFunc(dbFunc(deploymentID, "repository")).
		WithPublic(context.Generator.Repository()).
		Exec()

	if err != nil {
		return makeError(err)
	}

	// Webhook
	err = resources.SsCreator(context.Client.CreateWebhook).
		WithDbFunc(dbFunc(deploymentID, "webhook")).
		WithPublic(context.Generator.Webhook()).
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

	err := resources.SsPlaceholderCreator().WithDbFunc(dbFunc(id, "placeholder")).Exec()
	if err != nil {
		return makeError(err)
	}

	return nil
}

func Delete(id string, ownerID ...string) error {
	log.Println("deleting harbor for", id)

	makeError := func(err error) error {
		return fmt.Errorf("failed to delete harbor for deployment %s. details: %w", id, err)
	}

	context, err := NewContext(id, ownerID...)
	if err != nil {
		if errors.Is(err, base.DeploymentDeletedErr) {
			return nil
		}

		return makeError(err)
	}

	err = resources.SsDeleter(func(int) error { return nil }).
		WithResourceID(context.Deployment.Subsystems.Harbor.Webhook.ID).
		WithDbFunc(dbFunc(id, "webhook")).
		Exec()
	if err != nil {
		return makeError(err)
	}

	err = resources.SsDeleter(context.Client.DeleteRepository).
		WithResourceID(context.Deployment.Subsystems.Harbor.Repository.Name).
		WithDbFunc(dbFunc(id, "repository")).
		Exec()
	if err != nil {
		return makeError(err)
	}

	err = resources.SsDeleter(context.Client.DeleteRobot).
		WithResourceID(context.Deployment.Subsystems.Harbor.Robot.ID).
		WithDbFunc(dbFunc(id, "robot")).
		Exec()
	if err != nil {
		return makeError(err)
	}

	err = resources.SsDeleter(func(int) error { return nil }).
		WithResourceID(context.Deployment.Subsystems.Harbor.Project.ID).
		WithDbFunc(dbFunc(id, "project")).
		Exec()
	if err != nil {
		return makeError(err)
	}

	return nil
}

func Update(id string, params *deploymentModel.UpdateParams) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update harbor for deployment %s. details: %w", id, err)
	}

	context, err := NewContext(id)
	if err != nil {
		if errors.Is(err, base.DeploymentDeletedErr) {
			return nil
		}

		return makeError(err)
	}

	if context.Deployment.Subsystems.Harbor.Placeholder {
		log.Println("received update for harbor placeholder. skipping")
		return nil
	}

	if params.Name != nil {
		// updating the name requires moving the repository, since it is a persistent storage
		// we do this by creating a new repository with its "placeholder" being the first repository

		newRepository := context.Generator.Repository()
		oldRepository := context.Deployment.Subsystems.Harbor.Repository

		if oldRepository.Name != newRepository.Name &&
			service.Created(&context.Deployment.Subsystems.Harbor.Project) &&
			service.Created(&oldRepository) &&
			service.NotCreated(newRepository) {

			newRepository.Placeholder.RepositoryName = oldRepository.Name
			newRepository.Placeholder.ProjectName = context.Deployment.Subsystems.Harbor.Project.Name

			err = resources.SsCreator(context.Client.CreateRepository).
				WithDbFunc(dbFunc(id, "repository")).
				WithPublic(newRepository).
				Exec()
			if err != nil {
				return makeError(err)
			}

			err = resources.SsDeleter(context.Client.DeleteRepository).
				WithResourceID(oldRepository.Name).
				WithDbFunc(func(interface{}) error { return nil }).
				Exec()
			if err != nil {
				return makeError(err)
			}
		}

		newRobot := context.Generator.Robot()
		oldRobot := context.Deployment.Subsystems.Harbor.Robot

		if oldRobot.Name != newRobot.Name {
			err = resources.SsCreator(context.Client.CreateRobot).
				WithDbFunc(dbFunc(id, "robot")).
				WithPublic(newRobot).
				Exec()
			if err != nil {
				return makeError(err)
			}

			err = resources.SsDeleter(context.Client.DeleteRobot).
				WithResourceID(oldRobot.ID).
				WithDbFunc(func(interface{}) error { return nil }).
				Exec()
			if err != nil {
				return makeError(err)
			}
		}
	}

	return nil
}

func EnsureOwner(id, oldOwnerID string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update harbor owner for deployment %s. details: %w", id, err)
	}

	context, err := NewContext(id)
	if err != nil {
		if errors.Is(err, base.DeploymentDeletedErr) {
			return nil
		}

		return makeError(err)
	}

	oldOwnerContext, err := NewContext(id, oldOwnerID)
	if err != nil {
		if errors.Is(err, base.DeploymentDeletedErr) {
			return nil
		}

		return makeError(err)
	}

	if context.Deployment.Subsystems.Harbor.Placeholder {
		log.Println("received update for harbor placeholder. skipping")
		return nil
	}

	// the only manual work we need to do before triggering a repair is
	//  - ensure the harbor project exists
	//  - ensure the repository is copied to the new project

	err = resources.SsCreator(context.Client.CreateProject).
		WithDbFunc(dbFunc(id, "project")).
		WithPublic(context.Generator.Project()).
		Exec()
	if err != nil {
		return makeError(err)
	}

	newRepository := context.Generator.Repository()
	oldRepository := context.Deployment.Subsystems.Harbor.Repository

	if oldRepository.ID != newRepository.ID &&
		service.Created(&context.Deployment.Subsystems.Harbor.Project) &&
		service.Created(&oldRepository) &&
		service.NotCreated(newRepository) {

		newRepository.Placeholder.RepositoryName = oldRepository.Name
		newRepository.Placeholder.ProjectName = subsystemutils.GetPrefixedName(oldOwnerID)

		err = resources.SsCreator(context.Client.CreateRepository).
			WithDbFunc(dbFunc(id, "repository")).
			WithPublic(newRepository).
			Exec()
		if err != nil {
			return makeError(err)
		}

		err = resources.SsDeleter(oldOwnerContext.Client.DeleteRepository).
			WithResourceID(oldRepository.Name).
			WithDbFunc(func(interface{}) error { return nil }).
			Exec()
		if err != nil {
			return makeError(err)
		}
	}

	// remove the old resources
	err = resources.SsDeleter(oldOwnerContext.Client.DeleteRobot).
		WithResourceID(context.Deployment.Subsystems.Harbor.Robot.ID).
		WithDbFunc(dbFunc(id, "robot")).
		Exec()
	if err != nil {
		return makeError(err)
	}

	err = resources.SsDeleter(func(int) error { return nil }).
		WithResourceID(context.Deployment.Subsystems.Harbor.Webhook.ID).
		WithDbFunc(dbFunc(id, "webhook")).
		Exec()
	if err != nil {
		return makeError(err)
	}

	// trigger repair to ensure the other subsystems are updated
	err = Repair(id)
	if err != nil {
		return makeError(err)
	}

	return nil
}

func Repair(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to repair harbor %s. details: %w", id, err)
	}

	context, err := NewContext(id)
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
		).WithResourceID(project.ID).WithDbFunc(dbFunc(id, "project")).WithGenPublic(context.Generator.Project()).Exec()

		if err != nil {
			return makeError(err)
		}
	} else {
		err = resources.SsCreator(context.Client.CreateProject).
			WithDbFunc(dbFunc(id, "project")).
			WithPublic(context.Generator.Project()).
			Exec()

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
		).WithResourceID(robot.ID).WithDbFunc(dbFunc(id, "robot")).WithGenPublic(context.Generator.Robot()).Exec()

		if err != nil {
			return makeError(err)
		}
	} else {
		err = resources.SsCreator(context.Client.CreateRobot).
			WithDbFunc(dbFunc(id, "robot")).
			WithPublic(context.Generator.Robot()).
			Exec()
		if err != nil {
			return makeError(err)
		}
	}

	// don't repair the repository, since it can't be updated anyway
	// also <<NEVER>> call "DeleteRepository" here since it is the persistent storage for the deployment
	// if it is updated in the future to actually repair, the delete-func must be empty: func(string) error { return nil }
	if repo := &context.Deployment.Subsystems.Harbor.Repository; service.NotCreated(repo) {
		err = resources.SsCreator(context.Client.CreateRepository).
			WithDbFunc(dbFunc(id, "repository")).
			WithPublic(context.Generator.Repository()).
			Exec()

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
		).WithResourceID(hook.ID).WithDbFunc(dbFunc(id, "webhook")).WithGenPublic(context.Generator.Webhook()).Exec()

		if err != nil {
			return makeError(err)
		}
	} else {
		err = resources.SsCreator(context.Client.CreateWebhook).
			WithDbFunc(dbFunc(id, "webhook")).
			WithPublic(context.Generator.Webhook()).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

func dbFunc(id, key string) func(interface{}) error {
	return func(data interface{}) error {
		if data == nil {
			return deploymentModel.New().DeleteSubsystemByID(id, "harbor."+key)
		}
		return deploymentModel.New().UpdateSubsystemByID(id, "harbor."+key, data)
	}
}
