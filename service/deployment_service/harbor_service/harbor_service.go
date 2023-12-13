package harbor_service

import (
	"fmt"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/service"
	"go-deploy/service/deployment_service/client"
	"go-deploy/service/resources"
	"go-deploy/utils/subsystemutils"
	"log"
)

// Create sets up Harbor for the deployment.
//
// It creates a project, robot, repository and webhook associated with the deployment and returns an error if any.
func (c *Client) Create(id string, params *deploymentModel.CreateParams) error {
	log.Println("setting up harbor for", params.Name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup harbor for deployment %s. details: %w", params.Name, err)
	}

	_, hc, g, err := c.Get(OptsAll(id))
	if err != nil {
		return makeError(err)
	}

	// Project
	err = resources.SsCreator(hc.CreateProject).
		WithDbFunc(dbFunc(id, "project")).
		WithPublic(g.Project()).
		Exec()

	if err != nil {
		return makeError(err)
	}

	// Robot
	err = resources.SsCreator(hc.CreateRobot).
		WithDbFunc(dbFunc(id, "robot")).
		WithPublic(g.Robot()).
		Exec()

	if err != nil {
		return makeError(err)
	}

	// Repository
	err = resources.SsCreator(hc.CreateRepository).
		WithDbFunc(dbFunc(id, "repository")).
		WithPublic(g.Repository()).
		Exec()

	if err != nil {
		return makeError(err)
	}

	// Webhook
	err = resources.SsCreator(hc.CreateWebhook).
		WithDbFunc(dbFunc(id, "webhook")).
		WithPublic(g.Webhook()).
		Exec()

	if err != nil {
		return makeError(err)
	}

	return nil
}

// CreatePlaceholder sets up a placeholder harbor for the deployment.
//
// It does not create any resources in Harbor, and only creates a placeholder entry in the database.
func (c *Client) CreatePlaceholder(id string) error {
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

// Delete deletes the Harbor setup for the deployment.
//
// It deletes all the resources associated with the deployment and returns an error if any.
// This will remove any persistent storage associated with the deployment.
func (c *Client) Delete(id string) error {
	log.Println("deleting harbor for", id)

	makeError := func(err error) error {
		return fmt.Errorf("failed to delete harbor for deployment %s. details: %w", id, err)
	}

	d, hc, _, err := c.Get(OptsNoGenerator(id))
	if err != nil {
		return makeError(err)
	}

	err = resources.SsDeleter(func(int) error { return nil }).
		WithResourceID(d.Subsystems.Harbor.Webhook.ID).
		WithDbFunc(dbFunc(id, "webhook")).
		Exec()
	if err != nil {
		return makeError(err)
	}

	err = resources.SsDeleter(hc.DeleteRepository).
		WithResourceID(d.Subsystems.Harbor.Repository.Name).
		WithDbFunc(dbFunc(id, "repository")).
		Exec()
	if err != nil {
		return makeError(err)
	}

	err = resources.SsDeleter(hc.DeleteRobot).
		WithResourceID(d.Subsystems.Harbor.Robot.ID).
		WithDbFunc(dbFunc(id, "robot")).
		Exec()
	if err != nil {
		return makeError(err)
	}

	err = resources.SsDeleter(func(int) error { return nil }).
		WithResourceID(d.Subsystems.Harbor.Project.ID).
		WithDbFunc(dbFunc(id, "project")).
		Exec()
	if err != nil {
		return makeError(err)
	}

	return nil
}

// Update updates the Harbor setup for the deployment.
//
// It updates any of the resources associated with fields in the update params and returns an error if any.
func (c *Client) Update(id string, params *deploymentModel.UpdateParams) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update harbor for deployment %s. details: %w", id, err)
	}

	d, hc, g, err := c.Get(OptsAll(id))
	if err != nil {
		return makeError(err)
	}

	if d.Subsystems.Harbor.Placeholder {
		log.Println("received update for harbor placeholder. skipping")
		return nil
	}

	if params.Name != nil {
		// updating the name requires moving the repository, since it is a persistent storage
		// we do this by creating a new repository with its "placeholder" being the first repository

		newRepository := g.Repository()
		oldRepository := d.Subsystems.Harbor.Repository

		if oldRepository.Name != newRepository.Name &&
			service.Created(&d.Subsystems.Harbor.Project) &&
			service.Created(&oldRepository) &&
			service.NotCreated(newRepository) {

			newRepository.Placeholder.RepositoryName = oldRepository.Name
			newRepository.Placeholder.ProjectName = d.Subsystems.Harbor.Project.Name

			err = resources.SsCreator(hc.CreateRepository).
				WithDbFunc(dbFunc(id, "repository")).
				WithPublic(newRepository).
				Exec()
			if err != nil {
				return makeError(err)
			}

			err = resources.SsDeleter(hc.DeleteRepository).
				WithResourceID(oldRepository.Name).
				WithDbFunc(func(interface{}) error { return nil }).
				Exec()
			if err != nil {
				return makeError(err)
			}
		}

		newRobot := g.Robot()
		oldRobot := d.Subsystems.Harbor.Robot

		if oldRobot.Name != newRobot.Name {
			err = resources.SsCreator(hc.CreateRobot).
				WithDbFunc(dbFunc(id, "robot")).
				WithPublic(newRobot).
				Exec()
			if err != nil {
				return makeError(err)
			}

			err = resources.SsDeleter(hc.DeleteRobot).
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

// EnsureOwner ensures the owner of the Harbor setup for the deployment.
//
// If the owner of the deployment does match with the ID specified by WithUserID,
// it will update the Harbor setup to match the new owner.
//
// This will always trigger a call to Repair.
func (c *Client) EnsureOwner(id string, oldOwnerID string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update harbor owner for deployment %s. details: %w", id, err)
	}

	d, hc, g, err := c.Get(OptsAll(id))
	if err != nil {
		return makeError(err)
	}

	if d.Subsystems.Harbor.Placeholder {
		log.Println("received owner update for harbor placeholder. skipping")
		return nil
	}

	// Set up the client for the old owner
	oldD, oldHC, _, err := New(nil).Get(OptsNoGenerator(id, client.ExtraOpts{UserID: oldOwnerID}))
	if err != nil {
		return makeError(err)
	}

	// The only manual work we need to do before triggering a repair is
	//  - ensure the harbor project exists
	//  - ensure the repository is copied to the new project
	//  - remove the old resources

	err = resources.SsCreator(hc.CreateProject).
		WithDbFunc(dbFunc(id, "project")).
		WithPublic(g.Project()).
		Exec()
	if err != nil {
		return makeError(err)
	}

	newRepository := g.Repository()
	oldRepository := oldD.Subsystems.Harbor.Repository

	if oldRepository.ID != newRepository.ID &&
		service.Created(&d.Subsystems.Harbor.Project) &&
		service.Created(&oldRepository) &&
		service.NotCreated(newRepository) {

		newRepository.Placeholder.RepositoryName = oldRepository.Name
		newRepository.Placeholder.ProjectName = subsystemutils.GetPrefixedName(oldOwnerID)

		// Create a new repository
		err = resources.SsCreator(hc.CreateRepository).
			WithDbFunc(dbFunc(id, "repository")).
			WithPublic(newRepository).
			Exec()
		if err != nil {
			return makeError(err)
		}

		// Delete the old repository
		err = resources.SsDeleter(oldHC.DeleteRepository).
			WithResourceID(oldRepository.Name).
			WithDbFunc(func(interface{}) error { return nil }).
			Exec()
		if err != nil {
			return makeError(err)
		}
	}

	// Remove the old resources
	err = resources.SsDeleter(oldHC.DeleteRobot).
		WithResourceID(d.Subsystems.Harbor.Robot.ID).
		WithDbFunc(dbFunc(id, "robot")).
		Exec()
	if err != nil {
		return makeError(err)
	}

	err = resources.SsDeleter(func(int) error { return nil }).
		WithResourceID(d.Subsystems.Harbor.Webhook.ID).
		WithDbFunc(dbFunc(id, "webhook")).
		Exec()
	if err != nil {
		return makeError(err)
	}

	// Trigger repair to ensure everything is set up correctly
	err = c.Repair(id)
	if err != nil {
		return makeError(err)
	}

	return nil
}

// Repair repairs the Harbor setup for the deployment.
//
// It repairs any of the resources associated with the deployment and returns an error if any.
//
// Repositories are not repaired since they don't include an update function, and cannot be recreated
// since they are persistent storage for the deployment.
func (c *Client) Repair(id string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to repair harbor %s. details: %w", id, err)
	}

	d, hc, g, err := c.Get(OptsAll(id))
	if err != nil {
		return makeError(err)
	}

	if project := &d.Subsystems.Harbor.Project; service.Created(project) {
		err = resources.SsRepairer(
			hc.ReadProject,
			hc.CreateProject,
			hc.UpdateProject,
			func(int) error { return nil },
		).WithResourceID(project.ID).WithDbFunc(dbFunc(id, "project")).WithGenPublic(g.Project()).Exec()

		if err != nil {
			return makeError(err)
		}
	} else {
		err = resources.SsCreator(hc.CreateProject).
			WithDbFunc(dbFunc(id, "project")).
			WithPublic(g.Project()).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	if robot := &d.Subsystems.Harbor.Robot; service.Created(robot) {
		err = resources.SsRepairer(
			hc.ReadRobot,
			hc.CreateRobot,
			hc.UpdateRobot,
			hc.DeleteRobot,
		).WithResourceID(robot.ID).WithDbFunc(dbFunc(id, "robot")).WithGenPublic(g.Robot()).Exec()

		if err != nil {
			return makeError(err)
		}
	} else {
		err = resources.SsCreator(hc.CreateRobot).
			WithDbFunc(dbFunc(id, "robot")).
			WithPublic(g.Robot()).
			Exec()
		if err != nil {
			return makeError(err)
		}
	}

	// Don't repair the repository, since it can't be updated anyway.
	// Also, <<NEVER>> call "DeleteRepository" here since it is the persistent storage for the deployment.
	// If it is updated in the future to actually repair, the delete-func must be empty: func(string) error { return nil }.
	if repo := &d.Subsystems.Harbor.Repository; service.NotCreated(repo) {
		err = resources.SsCreator(hc.CreateRepository).
			WithDbFunc(dbFunc(id, "repository")).
			WithPublic(g.Repository()).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	if hook := &d.Subsystems.Harbor.Webhook; service.Created(hook) {
		err = resources.SsRepairer(
			hc.ReadWebhook,
			hc.CreateWebhook,
			hc.UpdateWebhook,
			hc.DeleteWebhook,
		).WithResourceID(hook.ID).WithDbFunc(dbFunc(id, "webhook")).WithGenPublic(g.Webhook()).Exec()

		if err != nil {
			return makeError(err)
		}
	} else {
		err = resources.SsCreator(hc.CreateWebhook).
			WithDbFunc(dbFunc(id, "webhook")).
			WithPublic(g.Webhook()).
			Exec()

		if err != nil {
			return makeError(err)
		}
	}

	return nil
}

// dbFunc returns a function that updates the Harbor subsystem.
func dbFunc(id, key string) func(interface{}) error {
	return func(data interface{}) error {
		if data == nil {
			return deploymentModel.New().DeleteSubsystemByID(id, "harbor."+key)
		}
		return deploymentModel.New().UpdateSubsystemByID(id, "harbor."+key, data)
	}
}
