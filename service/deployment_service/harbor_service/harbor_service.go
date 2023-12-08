package harbor_service

import (
	"fmt"
	deploymentModel "go-deploy/models/sys/deployment"
	"go-deploy/service"
	"go-deploy/service/resources"
	"go-deploy/utils/subsystemutils"
	"log"
)

func (c *Client) Create(params *deploymentModel.CreateParams) error {
	log.Println("setting up harbor for", params.Name)

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup harbor for deployment %s. details: %w", params.Name, err)
	}

	_, hc, g, err := c.Get(OptsNoDeployment)
	if err != nil {
		return makeError(err)
	}

	// Project
	err = resources.SsCreator(hc.CreateProject).
		WithDbFunc(dbFunc(c.ID(), "project")).
		WithPublic(g.Project()).
		Exec()

	if err != nil {
		return makeError(err)
	}

	// Robot
	err = resources.SsCreator(hc.CreateRobot).
		WithDbFunc(dbFunc(c.ID(), "robot")).
		WithPublic(g.Robot()).
		Exec()

	if err != nil {
		return makeError(err)
	}

	// Repository
	err = resources.SsCreator(hc.CreateRepository).
		WithDbFunc(dbFunc(c.ID(), "repository")).
		WithPublic(g.Repository()).
		Exec()

	if err != nil {
		return makeError(err)
	}

	// Webhook
	err = resources.SsCreator(hc.CreateWebhook).
		WithDbFunc(dbFunc(c.ID(), "webhook")).
		WithPublic(g.Webhook()).
		Exec()

	if err != nil {
		return makeError(err)
	}

	return nil
}

func (c *Client) CreatePlaceholder() error {
	log.Println("setting up placeholder harbor")

	makeError := func(err error) error {
		return fmt.Errorf("failed to setup placeholder harbor. details: %w", err)
	}

	err := resources.SsPlaceholderCreator().WithDbFunc(dbFunc(c.ID(), "placeholder")).Exec()
	if err != nil {
		return makeError(err)
	}

	return nil
}

func (c *Client) Delete() error {
	log.Println("deleting harbor for", c.ID())

	makeError := func(err error) error {
		return fmt.Errorf("failed to delete harbor for deployment %s. details: %w", c.ID(), err)
	}

	d, hc, _, err := c.Get(OptsNoGenerator)
	if err != nil {
		return makeError(err)
	}

	err = resources.SsDeleter(func(int) error { return nil }).
		WithResourceID(d.Subsystems.Harbor.Webhook.ID).
		WithDbFunc(dbFunc(c.ID(), "webhook")).
		Exec()
	if err != nil {
		return makeError(err)
	}

	err = resources.SsDeleter(hc.DeleteRepository).
		WithResourceID(d.Subsystems.Harbor.Repository.Name).
		WithDbFunc(dbFunc(c.ID(), "repository")).
		Exec()
	if err != nil {
		return makeError(err)
	}

	err = resources.SsDeleter(hc.DeleteRobot).
		WithResourceID(d.Subsystems.Harbor.Robot.ID).
		WithDbFunc(dbFunc(c.ID(), "robot")).
		Exec()
	if err != nil {
		return makeError(err)
	}

	err = resources.SsDeleter(func(int) error { return nil }).
		WithResourceID(d.Subsystems.Harbor.Project.ID).
		WithDbFunc(dbFunc(c.ID(), "project")).
		Exec()
	if err != nil {
		return makeError(err)
	}

	return nil
}

func (c *Client) Update(params *deploymentModel.UpdateParams) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update harbor for deployment %s. details: %w", c.ID(), err)
	}

	d, hc, g, err := c.Get(OptsAll)
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
				WithDbFunc(dbFunc(c.ID(), "repository")).
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
				WithDbFunc(dbFunc(c.ID(), "robot")).
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

func (c *Client) EnsureOwner(oldOwnerID string) error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to update harbor owner for deployment %s. details: %w", c.ID(), err)
	}

	d, hc, g, err := c.Get(OptsAll)
	if err != nil {
		return makeError(err)
	}

	if d.Subsystems.Harbor.Placeholder {
		log.Println("received owner update for harbor placeholder. skipping")
		return nil
	}

	// Set up the client for the old owner
	oldD, oldHC, _, err := New(nil).WithID(c.ID()).WithUserID(oldOwnerID).Get(OptsNoGenerator)
	if err != nil {
		return makeError(err)
	}

	// The only manual work we need to do before triggering a repair is
	//  - ensure the harbor project exists
	//  - ensure the repository is copied to the new project
	//  - remove the old resources

	err = resources.SsCreator(hc.CreateProject).
		WithDbFunc(dbFunc(c.ID(), "project")).
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
			WithDbFunc(dbFunc(c.ID(), "repository")).
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
		WithDbFunc(dbFunc(c.ID(), "robot")).
		Exec()
	if err != nil {
		return makeError(err)
	}

	err = resources.SsDeleter(func(int) error { return nil }).
		WithResourceID(d.Subsystems.Harbor.Webhook.ID).
		WithDbFunc(dbFunc(c.ID(), "webhook")).
		Exec()
	if err != nil {
		return makeError(err)
	}

	// Trigger repair to ensure everything is set up correctly
	err = c.Repair()
	if err != nil {
		return makeError(err)
	}

	return nil
}

func (c *Client) Repair() error {
	makeError := func(err error) error {
		return fmt.Errorf("failed to repair harbor %s. details: %w", c.ID(), err)
	}

	d, hc, g, err := c.Get(OptsAll)
	if err != nil {
		return makeError(err)
	}

	if project := &d.Subsystems.Harbor.Project; service.Created(project) {
		err = resources.SsRepairer(
			hc.ReadProject,
			hc.CreateProject,
			hc.UpdateProject,
			func(int) error { return nil },
		).WithResourceID(project.ID).WithDbFunc(dbFunc(c.ID(), "project")).WithGenPublic(g.Project()).Exec()

		if err != nil {
			return makeError(err)
		}
	} else {
		err = resources.SsCreator(hc.CreateProject).
			WithDbFunc(dbFunc(c.ID(), "project")).
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
		).WithResourceID(robot.ID).WithDbFunc(dbFunc(c.ID(), "robot")).WithGenPublic(g.Robot()).Exec()

		if err != nil {
			return makeError(err)
		}
	} else {
		err = resources.SsCreator(hc.CreateRobot).
			WithDbFunc(dbFunc(c.ID(), "robot")).
			WithPublic(g.Robot()).
			Exec()
		if err != nil {
			return makeError(err)
		}
	}

	// don't repair the repository, since it can't be updated anyway
	// also <<NEVER>> call "DeleteRepository" here since it is the persistent storage for the deployment
	// if it is updated in the future to actually repair, the delete-func must be empty: func(string) error { return nil }
	if repo := &d.Subsystems.Harbor.Repository; service.NotCreated(repo) {
		err = resources.SsCreator(hc.CreateRepository).
			WithDbFunc(dbFunc(c.ID(), "repository")).
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
		).WithResourceID(hook.ID).WithDbFunc(dbFunc(c.ID(), "webhook")).WithGenPublic(g.Webhook()).Exec()

		if err != nil {
			return makeError(err)
		}
	} else {
		err = resources.SsCreator(hc.CreateWebhook).
			WithDbFunc(dbFunc(c.ID(), "webhook")).
			WithPublic(g.Webhook()).
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
